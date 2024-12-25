package cache

import (
	"errors"
	"sync"
	"time"
)

type Cache struct {
	items     map[string]Item
	defaulTTL time.Duration
	mu        sync.RWMutex
}

type Item struct {
	Value   any
	Created time.Time
	Ttl     int64
}

// в кеше будет хранится n дней результат запроса на получение всех стендов или свободного стенда
// нужно хранить в кеше чтобы не делать кучу запросов в базу чтобы проверить какие стенды у нас есть и какие из них вообще свободные
// как только время кончится, кеш вызывает метод Clear()
func New(defaultTTL, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:     make(map[string]Item),
		defaulTTL: defaultTTL,
	}

	go c.collectGarbage(cleanupInterval)

	return c
}

func (c *Cache) WithTTL(key string, ttl time.Duration) {
	var expiration int64

	if ttl == 0 {
		ttl = c.defaulTTL
	}

	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Ttl: expiration,
	}
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Value:   value,
		Created: time.Now(),
	}
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, found := c.items[key]
	if found {
		delete(c.items, key)
	}

	return errors.New("item not found")
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.Ttl > 0 {
		if time.Now().UnixNano() > item.Ttl {
			return nil, false
		}
	}

	return item.Value, true
}

func (c *Cache) collectGarbage(cleanupInterval time.Duration) {
	for {
		<-time.After(cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

func (c *Cache) expiredKeys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0)

	for k, i := range c.items {
		if time.Now().UnixNano() > i.Ttl && i.Ttl > 0 {
			keys = append(keys, k)
		}
	}

	return keys
}

func (c *Cache) clearItems(keys []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, k := range keys {
		delete(c.items, k)
	}
}
