package ranking

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"testing"
)

var dataProvider = []struct {
	city  string
	score int
}{
	{"barcelona", 10},
	{"london", 6},
	{"madrid", 3},
	{"foo", 4},
	{"bar", 4},
	{"zoo", 4},
}

func TestRedisRankingPopulatesRequestedLocations(t *testing.T) {
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

	r := NewRedis(cl)
	defer func() {
		_, err := r.client.Del(context.Background(), SortedSetKey).Result()
		if err != nil {
			t.Fatalf("unexpected error removing sorted set, error %v", err)
		}
	}()

	for _, d := range dataProvider {
		for i := 0; i < d.score; i++ {
			err := r.IncreaseScore(d.city)
			if err != nil {
				t.Fatalf("unexpected error populating ranking, error %v", err)
			}
		}
	}

	size, err := r.Len()
	if err != nil {
		t.Fatalf("unexpected error getting length, error %v", err)
	}

	if int(size) != len(dataProvider) {
		t.Fatalf("expected sizes do not match, expcted %d got %d", len(dataProvider), size)
	}
}

func TestRedisRankingTopLocations(t *testing.T) {
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

	r := NewRedis(cl)
	defer func() {
		_, err := r.client.Del(context.Background(), SortedSetKey).Result()
		if err != nil {
			t.Fatalf("unexpected error removing sorted set, error %v", err)
		}
	}()

	for _, d := range dataProvider {
		for i := 0; i < d.score; i++ {
			err := r.IncreaseScore(d.city)
			if err != nil {
				t.Fatalf("unexpected error populating ranking, error %v", err)
			}
		}
	}

	topSize := 5
	res, err := r.Top(topSize)
	if err != nil {
		t.Fatalf("unexpected error getting top, error %v", err)
	}

	if len(res) != topSize {
		t.Errorf("unexpected top size, expected %d got %d", topSize, len(res))
	}

	expectedTop := "barcelona"
	if res[0].Name != expectedTop {
		t.Errorf("unexpected top location, expected %s got %s", expectedTop, res[0].Name)
	}

	expected := 10
	if res[0].Score != expected {
		t.Errorf("unexpected top score, expected %d got %d", expected, res[0].Score)
	}
}
