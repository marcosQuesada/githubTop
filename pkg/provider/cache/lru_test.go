package cache

import (
	"context"
	"testing"
	"time"
)

func TestAddEntriesToCacheAndExpireOldEntriesCleansCache(t *testing.T) {
	ttl := time.Millisecond * 100
	exp := time.Hour
	c, err := NewLRUCache(ttl, exp)
	if err != nil {
		t.Fatalf("Unexpected error creating cache, err: %s", err.Error())
	}
	defer c.Terminate()

	_ = c.Add(context.Background(), "key_1", "foo")
	_ = c.Add(context.Background(), "key_2", "bar")
	_ = c.Add(context.Background(), "key_3", "zzz")

	if c.Len() != 3 {
		t.Errorf("Unexpected cache size, expected 3 got %d", c.Len())
	}

	// Sleeps on tests are not the correct way to test things, but in this case
	// if something goes wrong, will take more time to execute, so entries will be expired
	time.Sleep(c.ttl)

	err = c.expire()
	if err != nil {
		t.Errorf("Unexpected error expiring cache, err: %s", err.Error())
	}

	if c.Len() != 0 {
		t.Errorf("Unexpected cache size, expected 0 got %d", c.Len())
	}
}

func TestAddEntriesToCacheAndExpirerCleansOldEntries(t *testing.T) {
	ttl := time.Millisecond * 100
	expWorker := time.Millisecond * 300
	c, err := NewLRUCache(ttl, expWorker)
	if err != nil {
		t.Fatalf("Unexpected error creating cache, err: %s", err.Error())
	}

	_ = c.Add(context.Background(), "key_1", "foo")
	_ = c.Add(context.Background(), "key_2", "bar")
	_ = c.Add(context.Background(), "key_3", "zzz")

	if c.Len() != 3 {
		t.Errorf("Unexpected cache size, expected 3 got %d", c.Len())
	}

	time.Sleep(time.Millisecond * 400)

	if c.Len() != 0 {
		t.Errorf("Unexpected cache size, expected 0 got %d", c.Len())
	}

	c.Terminate()
}
