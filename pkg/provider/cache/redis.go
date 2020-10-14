package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"time"
)

type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedis(cl *redis.Client, ttl time.Duration) *Redis {
	return &Redis{
		client: cl,
		ttl:    ttl,
	}
}

func (r *Redis) Add(k string, v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	cmd := r.client.Set(context.Background(), k, raw, r.ttl) //@TODO: Clean out context!

	return cmd.Err()
}

func (r *Redis) Get(k string) (interface{}, error) {
	cmd := r.client.Get(context.Background(), k)
	if cmd.Err() == redis.Nil {
		return nil,  provider.ErrCacheMiss
	}
	raw, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	var res []*provider.Contributor
	err =json.Unmarshal([]byte(raw), &res)

	return res, err
}

func (r *Redis) Terminate() {
	_ = r.client.Close()
}

