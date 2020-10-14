package provider

import "context"

// Location models searched locations
type Location struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Index int    `json:"index"`
}

type Ranking interface {
	IncreaseScore(city string) error
	Top(size int) ([]*Location, error)
}

type LocationRanking struct {
	ranking Ranking
}

func NewLocationRanking(r Ranking) *LocationRanking {
	return &LocationRanking{
		ranking: r,
	}
}

func (d *LocationRanking) IncreaseCityScore(_ context.Context, city string) error {
	return d.ranking.IncreaseScore(city)
}

func (d *LocationRanking) GetTopSearchedLocations(_ context.Context, size int) ([]*Location, error) {
	return d.ranking.Top(size)
}
