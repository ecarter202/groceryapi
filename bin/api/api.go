package api

import (
	"encoding/json"
	"log"
	"net/http"

	"grocery/config"
	"grocery/database"
	"grocery/models"
	"grocery/server"

	"github.com/gocraft/web"
)

const (
	_successfulMsg = "Success!"
)

type (
	GroceryAPI struct {
		*server.Context
		*server.Server
	}
)

func NewGroceryAPI() *server.Server {
	api := server.NewServer(config.APIPORT)

	server.Router.Subrouter(GroceryAPI{}, "/status").
		Get("/", (*GroceryAPI).Status)
	server.Router.Subrouter(GroceryAPI{}, "/products").
		Get("/search", (*GroceryAPI).Search).
		Get("/:id", (*GroceryAPI).Get).
		Post("/", (*GroceryAPI).Create).
		Delete("/:id", (*GroceryAPI).Delete)

	return api
}

func (api *GroceryAPI) Status(rw web.ResponseWriter, req *web.Request) {
	log.Print("checking status")

	api.Respond(rw, 200, "Running")
}

func (api *GroceryAPI) Search(rw web.ResponseWriter, req *web.Request) {
	log.Print("searching products")

	if keyword := req.URL.Query().Get("keyword"); keyword != "" {
		products := database.DB.Search(keyword)
		api.Respond(rw, http.StatusOK, _successfulMsg, products)
	} else {
		api.Respond(rw, http.StatusBadRequest, "invalid search")
	}
}

func (api *GroceryAPI) Get(rw web.ResponseWriter, req *web.Request) {
	if code, ok := req.PathParams["id"]; ok {
		if product := database.DB.Get(code); product != nil {
			api.Respond(rw, http.StatusOK, _successfulMsg, product)
		} else {
			api.Respond(rw, http.StatusNoContent, _successfulMsg)
		}
	} else {
		// NotFound middleware likely to hit before this is returned
		api.Respond(rw, http.StatusBadRequest, "invalid product code")
	}
}

func (api *GroceryAPI) Create(rw web.ResponseWriter, req *web.Request) {
	var products []*models.Product

	err := json.NewDecoder(req.Body).Decode(&products)
	if err != nil {
		api.Respond(rw, http.StatusBadRequest, "invalid product data")
		return
	}
	if len(products) == 0 {
		api.Respond(rw, http.StatusNoContent, "no product data supplied")
		return
	}

	createdProducts, errs := database.DB.Put(products...)
	if len(errs) > 0 {
		api.Respond(rw, http.StatusInternalServerError, "unable to create product", errs)
		log.Print("error creating products [ERR: ]", errs)
		return
	}

	api.Respond(rw, http.StatusOK, _successfulMsg, createdProducts)
}

func (api *GroceryAPI) Delete(rw web.ResponseWriter, req *web.Request) {
	if code, ok := req.PathParams["id"]; ok {
		database.DB.Del(code)
		api.Respond(rw, http.StatusOK, _successfulMsg)
		return
	}

	api.Respond(rw, http.StatusBadRequest, "invalid product code")
}
