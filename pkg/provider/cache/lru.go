package cache

import (
	"errors"
	"fmt"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"sync"
	"time"
)

const (
	// LruBaseSize defines LRU size
	LruBaseSize = 128
)

var (
	// ErrUnexpectedType happens on non expected type
	ErrUnexpectedType = errors.New("unexpected type")
)

// LruCache defines an LRU cache with expiration
type LruCache struct {
	ttl        time.Duration
	expireFreq time.Duration
	lru        simplelru.LRUCache
	mutex      sync.RWMutex
	done       chan struct{}
}

// Entry lRU entry
type Entry struct {
	Expire time.Time
	Value  interface{}
}

// NewLRUCache instantiates LRU cache
func NewLRUCache(ttl, freq time.Duration) (*LruCache, error) {
	lru, err := simplelru.NewLRU(LruBaseSize, func(key interface{}, value interface{}) {})
	if err != nil {
		return nil, err
	}

	l := &LruCache{
		lru:        lru,
		ttl:        ttl,
		expireFreq: freq,
		done:       make(chan struct{}),
	}
	go l.runner()

	return l, nil
}

// Add LRU entry
func (c *LruCache) Add(k string, v interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := newEntry(c.ttl, v)
	c.lru.Add(k, entry)

	return nil
}

// Get LRU entry
func (c *LruCache) Get(k string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	v, ok := c.lru.Get(k)
	if !ok {
		return nil, provider.ErrCacheMiss
	}

	vv, ok := v.(*Entry)
	if !ok {
		return nil, ErrUnexpectedType
	}

	return vv.Value, nil
}

// Len returns cache size
func (c *LruCache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.lru.Len()
}

// Terminate stop worker
func (c *LruCache) Terminate() {
	close(c.done)
}

// nolint:unused
func (c *LruCache) remove(k string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ok := c.lru.Remove(k)
	if !ok {
		return fmt.Errorf("Entry %s not removed", k)
	}

	return nil
}

func (c *LruCache) runner() {
	ticker := time.NewTicker(c.expireFreq)
	for {
		select {
		case <-ticker.C:
			err := c.expire()
			if err != nil {
				log.Errorf("Unexpected error expiring cache, err: %s", err.Error())
			}
		case <-c.done:
			ticker.Stop()

			return
		}
	}
}

func (c *LruCache) expire() error {
	v, isFound := c.getOldest()
	if !isFound {
		// Cache is empty, do nothing
		return nil
	}

	if v.(*Entry).Expire.Unix() <= time.Now().Unix() {
		c.removeOldest()

		//Check more entries for expiration
		if c.Len() > 0 {
			return c.expire()
		}
	}

	return nil
}

func (c *LruCache) getOldest() (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, v, isFound := c.lru.GetOldest()

	return v, isFound
}

func (c *LruCache) removeOldest() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lru.RemoveOldest()
}

func newEntry(ttl time.Duration, v interface{}) *Entry {
	return &Entry{
		Expire: time.Now().Local().Add(ttl),
		Value:  v,
	}
}
