package mpesa

import (
	"sync"
	"time"
)

type Cache struct {
	data map[string]*AccessTokenResponse
	lock *sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*AccessTokenResponse),
		lock: &sync.RWMutex{},
	}
}

// Get retrieves the token from cache.
func (c *Cache) Get(key string) (*AccessTokenResponse, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if time.Until(v.ExpireTime).Minutes() <= 0 {
		return nil, false
	}
	return v, true
}

// Set Adds the token to cache.
func (c *Cache) Set(val *AccessTokenResponse) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data["token"] = val
}
