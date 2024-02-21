package main

import (
	"flag"
	"log"
	"os/signal"
	"syscall"

	"api"
	"grocery/database"
	"grocery/shared"
)

var (
	_debug bool
)

func main() {
	flag.BoolVar(&_debug, "debug", false, "Enable debugging.")
	flag.Parse()

	if _debug {
		shared.SetDebug()
	}

	s := api.NewGroceryAPI()
	go s.Run()

	database.Connect()

	//setup signal handling to respond to ctrl-c
	signal.Notify(
		shared.SigChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	sig := <-shared.SigChannel
	log.Printf("caught signal %q", sig)
	close(shared.ShutdownChan)

	s.ShutDown()
	log.Println("DONE")
}
