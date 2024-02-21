package database

import (
	"grocery/models"
	"testing"
)

func TestConnect(t *testing.T) {
	db := Connect()
	if db == nil || DB == nil {
		t.Error("database failed to create/connect")
	}
}

func TestSearch(t *testing.T) {
	db := Connect()
	items := db.Search("e")
	if len(items) < 1 {
		t.Error("search failed to locate any dummy data items")
	}
}

func TestGet(t *testing.T) {
	db := Connect()
	item := db.Get("A12T-4GH7-QPL9-3N4M")
	if item == nil {
		t.Error("failed to get database item")
	}
}

func TestPut(t *testing.T) {
	wantedPrice := 9.81

	db := Connect()
	items, errs := db.Put(&models.Product{
		Name:  "Waffles",
		Price: 9.8111111,
	})
	if len(errs) > 0 {
		t.Errorf("errors when creating product(s) [ERR: %s]", errs)
	}
	if len(items) == 0 {
		t.Errorf("failed to create product [ERRS: %s]", errs)
	} else if items[0].Price != wantedPrice {
		t.Errorf("failed to round price during creation; wanted %v but got %v", wantedPrice, items[0].Price)
	}
}

func TestDel(t *testing.T) {
	db := Connect()
	initialDBSize := len(db.Items)

	db.Del("A12T-4GH7-QPL9-3N4M")

	if len(db.Items) >= initialDBSize {
		t.Error("failed to delete item from database")
	}
}
