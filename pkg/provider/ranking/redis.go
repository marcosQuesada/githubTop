package ranking

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/marcosQuesada/githubTop/pkg/provider"
)

const SortedSetKey = "location-ranking"

// Redis implements a redis baked ranking in top of a sorted set
type Redis struct {
	client *redis.Client
}

// NewRedis instantiates redis ranking
func NewRedis(cl *redis.Client) *Redis {
	return &Redis{
		client: cl,
	}
}

// IncreaseScore city score  increase by 1
func (r *Redis) IncreaseScore(city string) error {
	return r.client.ZIncr(context.Background(), SortedSetKey, &redis.Z{
		Score:  1,
		Member: city,
	}).Err()

}

// Top returns priority queue from head up to "size" length
func (r *Redis) Top(size int) ([]*provider.Location, error) {
	res, err := r.client.ZRevRangeByScoreWithScores(context.Background(), SortedSetKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, err
	}

	var top []*provider.Location
	if len(res) > size {
		res = res[:size]
	}

	for i, re := range res {
		top = append(top, &provider.Location{
			Name:  re.Member.(string),
			Score: int(re.Score),
			Index: i,
		})
	}

	return top, nil
}

// Len returns ranking size
func (r *Redis) Len() (int64, error) {
	return r.client.ZCount(context.Background(), SortedSetKey, "-inf", "+inf").Result()
}
