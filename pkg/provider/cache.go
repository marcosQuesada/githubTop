package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/marcosQuesada/githubTop/pkg/log"

	"time"
)

var(
	ErrCacheMiss      = errors.New("entry not found in cache.")
)

type Cache interface {
	// Add entry to cache
	Add(k string, v interface{}) error

	// Get entry from cache, puts Entry in top of LRU, so refresh expiration entry
	Get(k string) (interface{}, error)

	// Stop expiration worker
	Terminate()
}

type cacheMiddleware struct {
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

// NewCacheMiddleware instantiates cached repository
func NewCacheMiddleware(cache Cache, repo GithubRepository) *cacheMiddleware {
	return &cacheMiddleware{
		cache:      cache,
		repository: repo,
	}
}

// GetGithubTopContributors tries cache lookup, on miss access repository
func (r *cacheMiddleware) GetGithubTopContributors(ctx context.Context, city string, size int) ([]*Contributor, error) {
	k := r.key(city, size)
	res, err := r.cache.Get(k)
	if err == nil {
		c, ok := res.([]*Contributor)
		if !ok {
			return nil, fmt.Errorf("unexpected cache type entry, type %T", res)
		}

		return c, nil
	}

	if err != ErrCacheMiss {
		//on unexpected cache errors, track it and let repository do its work
		log.Errorf("Unexpected Error reading cache, err: %s", err.Error())
	}

	c, errc := r.repository.GetGithubTopContributors(ctx, city, size)
	if errc != nil {
		return nil, errc
	}

	if err = r.AddTopContributors(city, size, c); err != nil {
		log.Errorf("Error adding element on cache is: %s", err.Error())
	}

	return c, nil
}

// AddTopContributors updates cache layer
func (r *cacheMiddleware) AddTopContributors(city string, size int, contributors []*Contributor) error {
	k := r.key(city, size)

	return r.cache.Add(k,  contributors)
}

// Terminate close repository
func (r *cacheMiddleware) Terminate() {
	r.cache.Terminate()
}

func (r *cacheMiddleware) key(city string, size int) string {
	return fmt.Sprintf("city_%s_size_%d", city, size)
}
