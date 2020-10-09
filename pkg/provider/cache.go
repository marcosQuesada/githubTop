package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"sync"
	"time"
)

const (
	// LRU_BASE_SIZE defines LRU size
	LRU_BASE_SIZE = 128
)

var (
	ERR_CACHE_MISS  = errors.New("Entry not found in cache.")
	UNEXPECTED_TYPE = errors.New("Unexpected type.")
)

type Cache interface {
	// Starts worker that purge expired entries
	RunExpirationWorker()

	// Stop expiration worker
	TerminateExpirationWorker()

	// Add entry to cache
	Add(k string, v interface{}) error

	// Get entry from cache, puts Entry in top of LRU, so refresh expiration entry
	Get(k string) (interface{}, error)

	// Get cache size
	Len() int
}

type defaultGithubCacheRepository struct {
	cache      Cache
	repository GithubRepository
}

// CacheConfig cache parameters configuration
type CacheConfig struct {
	Ttl                 time.Duration
	ExpirationFrequency time.Duration
}

// NewCacheConfig instantiates config
func NewCacheConfig(ttl, exp time.Duration) CacheConfig {
	return CacheConfig{ttl, exp}
}

// NewGithubRepositoryCache instantiates cached repository
func NewGithubRepositoryCache(cfg CacheConfig, repo GithubRepository) *defaultGithubCacheRepository {
	c, _ := NewLRUCache(cfg.Ttl, cfg.ExpirationFrequency)
	c.RunExpirationWorker()

	return &defaultGithubCacheRepository{
		cache:      c,
		repository: repo,
	}
}

// GetGithubTopContributors tries cache lookup, on miss access repository
func (r *defaultGithubCacheRepository) GetGithubTopContributors(ctx context.Context, city string, size int) ([]*Contributor, error) {
	k := r.key(city, size)
	res, err := r.cache.Get(k)
	if err == nil {
		c, ok := res.([]*Contributor)
		if !ok {
			return nil, fmt.Errorf("Unexpected cache type entry, type %v", res)
		}

		return c, nil
	}

	if err != ERR_CACHE_MISS {
		//on unexpected cache errors, track it and let repository do its work
		log.Errorf("Unexpected Error reading cache, err: %s", err.Error())
	}

	c, err := r.repository.GetGithubTopContributors(ctx, city, size)
	if err != nil {
		return nil, err
	}

	err = r.AddTopContributors(city, size, c)
	if err != nil {
		log.Infof("Error adding element on cache is: %s", err.Error())
	}

	return c, nil
}

// AddTopContributors updates cache layer
func (r *defaultGithubCacheRepository) AddTopContributors(city string, size int, contributors []*Contributor) error {
	k := r.key(city, size)

	return r.cache.Add(k, contributors)
}

// Terminate close repository
func (r *defaultGithubCacheRepository) Terminate() {
	r.cache.TerminateExpirationWorker()
}

func (r *defaultGithubCacheRepository) key(city string, size int) string {
	return fmt.Sprintf("city_%s_size_%d", city, size)
}

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
func NewLRUCache(ttl, freq time.Duration) (c *LruCache, err error) {
	lru, err := simplelru.NewLRU(LRU_BASE_SIZE, func(key interface{}, value interface{}) {})
	if err != nil {
		return nil, err
	}

	return &LruCache{
		lru:        lru,
		ttl:        ttl,
		expireFreq: freq,
		done:       make(chan struct{}),
	}, nil
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
		return nil, ERR_CACHE_MISS
	}

	vv, ok := v.(*Entry)
	if !ok {
		return nil, UNEXPECTED_TYPE
	}

	return vv.Value, nil
}

// Len returns cache size
func (c *LruCache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.lru.Len()
}

// RunExpirationWorker runs expiration worker
func (c *LruCache) RunExpirationWorker() {
	go c.runner()
}

// TerminateExpirationWorker stop worker
func (c *LruCache) TerminateExpirationWorker() {
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
