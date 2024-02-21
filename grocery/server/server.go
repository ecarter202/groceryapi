package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"grocery/cache"
	logger "grocery/log"
	"grocery/sec"
	"grocery/shared"

	"github.com/gocraft/web"
)

var (
	_maxConnections = 100
	_memoryLimit    = 10 << 20
	_maxRPS         = 15.0
	_connChan       = make(chan int, _maxConnections)

	rateCache = cache.NewCache()
	Router    *web.Router
)

type (
	Server struct {
		*logger.Logger
		*http.Server
	}

	Context struct {
		*log.Logger `json:"-"`

		ReqStartTime time.Time `json:"-"`
		Body         []byte    `json:"-"`
	}
)

func NewServer(port int) *Server {
	s := &Server{
		Logger: logger.NewLogger(
			fmt.Sprintf("[serve-%s] ", shared.MODE),
			log.Lmsgprefix,
		),
	}

	for i := 0; i < _maxConnections; i++ {
		_connChan <- i
	}

	Router = web.New(Context{}).
		Middleware((*Context).InitLogger).
		Middleware((*Context).InitStartTime).
		Middleware((*Context).RateLimit).
		NotFound((*Context).NotFound).
		OptionsHandler((*Context).OptionsHandler)

	s.Server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           http.MaxBytesHandler(Router, int64(_memoryLimit)),
		ReadHeaderTimeout: 2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
	}

	return s
}

func (ctx *Context) InitStartTime(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	ctx.ReqStartTime = time.Now()
	next(rw, req)
}

// InitLogger sets up the logger for the server context.
func (ctx *Context) InitLogger(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	ctx.Logger = log.New(
		os.Stdout,
		fmt.Sprintf("%s \"%s\" ", shared.MODE, req.URL.Path),
		log.Lmsgprefix,
	)

	next(rw, req)
}

// RateLimit the API requests
func (ctx *Context) RateLimit(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	select {
	case <-_connChan:
	default:
		ctx.Respond(rw, http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}

	// add connect back
	defer func() {
		_connChan <- 1
	}()

	var (
		rps      float64
		hash     string
		clientIP string
	)

	if len(req.RemoteAddr) > 0 {
		clientIP = req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")]
	}

	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		hash = shared.HashSha1(fmt.Sprintf("%s_%s", clientIP, xff))
	} else {
		hash = shared.HashSha1(clientIP)
	}

	if !rateCache.Has(hash) {
		_ = rateCache.Put(hash)
	} else {
		created, _, counter := rateCache.Inc(hash)

		//get the total length of time this ip has been making requests
		seenSecs := ctx.ReqStartTime.Sub(created).Seconds()
		if seenSecs > 1.0 {
			rps = counter / seenSecs
			if rps >= _maxRPS {
				ctx.Respond(rw, http.StatusTooManyRequests, "Request rate limit exceeded")
				return
			}
		}

	}

	next(rw, req)
}

func (ctx *Context) OptionsHandler(rw web.ResponseWriter, req *web.Request, methods []string) {
	rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	rw.Header().Set("Access-Control-Max-Age", "86400")
}

func (ctx *Context) SetHeaders(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	rw.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	rw.Header().Set("Access-Control-Allow-Credentials", "true")

	next(rw, req)
}

func (ctx *Context) NotFound(rw web.ResponseWriter, req *web.Request) {
	ctx.Respond(rw, 404, req.RequestURI)
}

func (s *Server) Run() {
	if shared.MODE == shared.MODE_DEBUG {
		s.Print("starting up server on \"%s\"", s.Server.Addr)
		s.Server.ListenAndServe()
	} else {
		s.Print("starting up SSL server on \"%s\"", s.Server.Addr)
		if err := listenAndServeTLS(s.Server, sec.S3, sec.S4); err != nil {
			s.Print("could not start server [ERR:%s]\n", err.Error())
			return
		}
	}

	s.Print("exiting server")
}

func (s *Server) ShutDown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Server.Shutdown(ctx)
}

// Respond will respond to the request
func (ctx *Context) Respond(rw web.ResponseWriter, code int, message string, data ...interface{}) {
	msg := &Message{
		Code:    code,
		Message: message,
	}

	if len(data) > 0 {
		if len(data) > 1 {
			msg.Data = data
		} else {
			msg.Data = data[0]
		}
	}

	buf := bytes.Buffer{}
	gz := gzip.NewWriter(&buf)
	gz.Write(msg.Marshal())
	gz.Close()

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Content-Encoding", "gzip")
	rw.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

	rw.WriteHeader(code)
	rw.Write(buf.Bytes())
}

func listenAndServeTLS(srv *http.Server, certPEMBlock, keyPEMBlock []byte) error {
	if srv.Addr == "" {
		srv.Addr = ":https"
	}

	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}

	if srv.TLSConfig.NextProtos == nil {
		srv.TLSConfig.NextProtos = []string{"http/1.1"}
	}

	var err error
	srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
	srv.TLSConfig.Certificates[0], err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, srv.TLSConfig)

	return srv.Serve(tlsListener)
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(30 * time.Second)
	return tc, nil
}
