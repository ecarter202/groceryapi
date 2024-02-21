package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"grocery/config"
	"grocery/database"
	"grocery/server"
)

func testAPISetup() {
	if server.Router == nil {
		database.Connect()

		server.NewServer(config.APIPORT)

		server.Router.Subrouter(GroceryAPI{}, "/status").
			Get("/", (*GroceryAPI).Status)
		server.Router.Subrouter(GroceryAPI{}, "/products").
			Get("/search", (*GroceryAPI).Search).
			Get("/:id", (*GroceryAPI).Get).
			Post("/", (*GroceryAPI).Create).
			Delete("/:id", (*GroceryAPI).Delete)
	}
}

func TestStatus(t *testing.T) {
	testAPISetup()

	req, err := http.NewRequest(http.MethodGet, "/status", nil)
	if err != nil {
		t.Fatalf("failed to create the request: %v\n", err)
	}

	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d but got %d\n", http.StatusOK, w.Code)
	}
}

func TestSearch(t *testing.T) {
	testAPISetup()

	values := url.Values{}
	values.Add("keyword", "e")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(
		"/products/search?%s",
		values.Encode(),
	), nil)
	if err != nil {
		t.Fatalf("failed to create the request: %v\n", err)
	}

	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)
	respBody := w.Result().Body

	reader, err := gzip.NewReader(respBody)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	var msg *server.Message

	err = json.Unmarshal(body, &msg)
	if err != nil {
		t.Fatalf("failed decoding response body [ERR: %s]", err)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d but got %d\n", http.StatusOK, w.Code)
	}

}

func TestGet(t *testing.T) {
	testAPISetup()

	var getTable = map[int]string{
		http.StatusOK:        database.DummyData[0].Code,
		http.StatusNoContent: "this-isnt-real-code",
	}

	for expectedStatusCode, productCode := range getTable {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(
			"/products/%s",
			productCode,
		), nil)
		if err != nil {
			t.Fatalf("failed to create the request: %v\n", err)
		}

		w := httptest.NewRecorder()

		server.Router.ServeHTTP(w, req)
		respBody := w.Result().Body

		reader, err := gzip.NewReader(respBody)
		if err != nil {
			t.Fatalf("failed to read gzip body [ERR: %s]", err)
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			reader.Close()
			t.Fatalf("failed to read gzip reader [ERR: %s]", err)
		}
		reader.Close()

		var msg *server.Message

		err = json.Unmarshal(body, &msg)
		if err != nil {
			t.Fatalf("failed decoding response body [ERR: %s]", err)
		}

		if w.Code != expectedStatusCode {
			t.Fatalf("expected status code %d but got %d\n", expectedStatusCode, w.Code)
		}
	}
}

func TestCreate(t *testing.T) {
	testAPISetup()

	var createTable = map[int][]byte{}
	dummydata, _ := json.Marshal(database.DummyData)
	createTable[http.StatusOK] = dummydata
	createTable[http.StatusNoContent] = []byte("[]")
	createTable[http.StatusBadRequest] = []byte(`{
		"menuitem": [
		  {"value": "New", "onclick": "CreateNewDoc()"},
		  {"value": "Open", "onclick": "OpenDoc()"},
		  {"value": "Close", "onclick": "CloseDoc()"}
		]
	  }`)

	for expectedStatusCode, productData := range createTable {
		req, err := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(productData))
		if err != nil {
			t.Fatalf("failed to create the request: %v\n", err)
		}

		w := httptest.NewRecorder()

		server.Router.ServeHTTP(w, req)
		respBody := w.Result().Body

		reader, err := gzip.NewReader(respBody)
		if err != nil {
			t.Fatalf("failed to read gzip body [ERR: %s]", err)
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			reader.Close()
			t.Fatalf("failed to read gzip reader [ERR: %s]", err)
		}
		reader.Close()

		var msg *server.Message

		err = json.Unmarshal(body, &msg)
		if err != nil {
			t.Fatalf("failed decoding response body [ERR: %s]", err)
		}

		if w.Code != expectedStatusCode {
			t.Fatalf("expected status code %d but got %d\n", expectedStatusCode, w.Code)
		}
	}
}

func TestDelete(t *testing.T) {
	testAPISetup()

	var delTable = map[int]string{
		http.StatusOK: database.DummyData[0].Code,
	}

	for expectedStatusCode, productCode := range delTable {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(
			"/products/%s",
			productCode,
		), nil)
		if err != nil {
			t.Fatalf("failed to create the request: %v\n", err)
		}

		w := httptest.NewRecorder()

		server.Router.ServeHTTP(w, req)
		respBody := w.Result().Body

		reader, err := gzip.NewReader(respBody)
		if err != nil {
			t.Fatalf("failed to read gzip body [ERR: %s]", err)
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			reader.Close()
			t.Fatalf("failed to read gzip reader [ERR: %s]", err)
		}
		reader.Close()

		var msg *server.Message

		err = json.Unmarshal(body, &msg)
		if err != nil {
			t.Fatalf("failed decoding response body [ERR: %s]", err)
		}

		if w.Code != expectedStatusCode {
			t.Fatalf("expected status code %d but got %d\n", expectedStatusCode, w.Code)
		}
	}
}
