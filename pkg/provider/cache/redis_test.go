package cache

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"testing"
	"time"
)

// @TODO: Mark as integration test!
func TestAddEntryOnRedisCache(t *testing.T) {
	if testing.Short() {
		log.Info("Skipping tests because of Short flag")
		return
	}

	opt := &redis.Options{
		Network: "",
		Addr:    ":6379",
		DB:      0,
	}
	cl := redis.NewClient(opt)
	r := NewRedis(cl, time.Second)

	key := "foo"
	value := []*provider.Contributor{
		{
			ID:   123,
			Name: "fooBar",
		},
	}
	err := r.Add(context.Background(), key, value)
	if err != nil {
		t.Fatalf("unexpected error adding cache entry, error %v", err)
	}

	res, err := r.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected error getting cache entry, error %v", err)
	}

	v, ok := res.([]*provider.Contributor)
	if !ok {
		t.Fatalf("unexpected type on cache, got %T", res)
	}

	if len(v) == 0 {
		t.Fatal("Unexpected cache response, empty size")
	}

	if v[0].Name != value[0].Name {
		t.Errorf("expected values do not match, expected %s got %s", v[0].Name, value[0].Name)
	}
}

func TestAddEntryOnRedisCacheAndWaitExpiration(t *testing.T) {
	if testing.Short() {
		log.Info("Skipping tests because of Short flag")
		return
	}

	opt := &redis.Options{
		Network: "",
		Addr:    ":6379",
		DB:      0,
	}
	cl := redis.NewClient(opt)
	r := NewRedis(cl, time.Second)

	key := "foo"
	value := []*provider.Contributor{
		{
			ID:   123,
			Name: "fooBar",
		},
	}
	err := r.Add(context.Background(), key, value)
	if err != nil {
		t.Fatalf("unexpected error adding cache entry, error %v", err)
	}

	time.Sleep(time.Second * 2)
	_, err = r.Get(context.Background(), key)
	if err == nil {
		t.Fatalf("unexpected error getting cache entry, error %v", err)
	}

	if !errors.Is(err, provider.ErrCacheMiss) {
		t.Errorf("unexpected error type, got %t", err)
	}
}
