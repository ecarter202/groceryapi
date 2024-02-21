package database

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"grocery/models"
	"grocery/shared"
)

var DB *Database

type (
	Database struct {
		sync.RWMutex

		Items []*models.Product
	}
)

func Connect() *Database {
	if DB == nil {
		DB = new(Database)
		loadDummyData(DB)
	}

	return DB
}

func (d *Database) Search(name string) (found []*models.Product) {
	d.RLock()

	for _, item := range d.Items {
		iName := strings.ToLower(item.Name)
		name = strings.ToLower(name)
		if strings.Contains(iName, name) {
			found = append(found, item)
		}
	}

	d.RUnlock()

	return
}

func (d *Database) Get(code string) *models.Product {
	if code == "" {
		return nil
	}

	d.RLock()
	defer d.RUnlock()

	for _, item := range d.Items {
		if strings.EqualFold(item.Code, code) {
			return item
		}
	}

	return nil
}

func (d *Database) Put(items ...*models.Product) (products []*models.Product, errs []error) {
	for _, item := range items {
		if !shared.IsAlphaNum(item.Name) {
			errs = append(errs, fmt.Errorf("name %q is not alphanumeric", item.Name))
		}
		if len(errs) > 0 {
			return
		}

		item.Code = shared.GenProductCode()
		item.Price = shared.RoundFloat(item.Price, 2)
	}

	d.Lock()
	d.Items = append(d.Items, items...)
	d.Unlock()

	return items, nil
}

func (d *Database) Del(code string) (err error) {
	if code == "" {
		return errors.New("invalid product code")
	}

	for i, item := range d.Items {
		if strings.EqualFold(item.Code, code) {
			d.Lock()
			d.Items = append(d.Items[:i], d.Items[i+1:]...)
			d.Unlock()
		}
	}

	return nil
}

func loadDummyData(d *Database) {
	log.Print("loading dummy data...")
	defer func() {
		log.Print("dummy data loaded")
	}()

	d.Items = append(d.Items, &models.Product{"A12T-4GH7-QPL9-3N4M", "Lettuce", 3.46})
	d.Items = append(d.Items, &models.Product{"E5T6-9UI3-TH15-QR88", "Peach", 2.99})
	d.Items = append(d.Items, &models.Product{"YRT6-72AS-K736-L4AR", "Green Pepper", 0.79})
	d.Items = append(d.Items, &models.Product{"TQ4C-VV6T-75ZX-1RMR", "Gala Apple", 3.59})
}
