package ranking

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/marcosQuesada/githubTop/pkg/provider"
)

const SortedSetKey = "location-ranking"

type Redis struct {
	client *redis.Client
}

func NewRedis(cl *redis.Client) *Redis {
	return &Redis{
		client: cl,
	}
}

func (r *Redis) IncreaseScore(city string) error {
	return r.client.ZIncr(context.Background(), SortedSetKey, &redis.Z{
		Score:  1,
		Member: city,
	}).Err()

}

func (r *Redis) Top(size int) ([]*provider.Location, error) {
	res, err := r.client.ZRevRangeByScoreWithScores(context.Background(), SortedSetKey,  &redis.ZRangeBy{
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

func(r *Redis) Len() (int64, error) {
	return r.client.ZCount(context.Background(), SortedSetKey, "-inf", "+inf").Result()
}