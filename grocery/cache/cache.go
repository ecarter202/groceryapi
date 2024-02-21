package cache

import (
	"errors"
	"sync"
	"time"
)

type (
	CacheItem struct {
		Data     interface{}
		count    float64
		created  time.Time
		accessed time.Time
	}

	Cache struct {
		sync.RWMutex

		Items map[string]*CacheItem
	}
)

func NewCache() *Cache {
	return &Cache{
		Items: make(map[string]*CacheItem),
	}
}

func (c *Cache) Has(key string) bool {
	c.RLock()
	_, ok := c.Items[key]
	c.RUnlock()

	return ok
}

func (c *Cache) Get(key string) (created, accessed time.Time, count float64) {
	if key == "" {
		return
	}

	c.RLock()
	defer c.RUnlock()

	if item, ok := c.Items[key]; ok {
		return item.created, item.accessed, item.count
	}

	return
}

func (c *Cache) Put(key string) (err error) {
	if key == "" {
		return errors.New("key required")
	}

	c.Lock()

	if _, ok := c.Items[key]; !ok {
		c.Items[key] = &CacheItem{created: time.Now()}
	}

	c.Items[key].accessed = time.Now()

	c.Unlock()

	return
}

func (c *Cache) Inc(key string) (created, accessed time.Time, count float64) {
	if key == "" {
		return
	}

	c.RLock()
	defer c.RUnlock()

	if item, ok := c.Items[key]; ok {
		now := time.Now()
		c.Items[key].accessed = now
		c.Items[key].count++

		return item.created, now, c.Items[key].count
	}

	return
}

func (c *Cache) Dec(key string) (created, accessed time.Time, count float64) {
	if key == "" {
		return
	}

	c.RLock()
	defer c.RUnlock()

	if item, ok := c.Items[key]; ok {
		now := time.Now()
		c.Items[key].accessed = now
		c.Items[key].count--

		return item.created, now, c.Items[key].count
	}

	return
}

func (c *Cache) Del(key string) {
	c.Lock()
	delete(c.Items, key)
	c.Unlock()
}
