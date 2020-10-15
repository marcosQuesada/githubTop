package provider

import "context"

// Location models searched locations
type Location struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Index int    `json:"index"`
}

// Ranking defines a generic Ranking
type Ranking interface {
	IncreaseScore(city string) error
	Top(size int) ([]*Location, error)
}

// LocationRanking defines top searched locations generic ranking
type LocationRanking struct {
	ranking Ranking
}

// NewLocationRanking wraps ranking persistence
func NewLocationRanking(r Ranking) *LocationRanking {
	return &LocationRanking{
		ranking: r,
	}
}

// IncreaseCityScore increase city score by 1
func (d *LocationRanking) IncreaseCityScore(_ context.Context, city string) error {
	return d.ranking.IncreaseScore(city)
}

// GetTopSearchedLocations returns ranking top with "size"
func (d *LocationRanking) GetTopSearchedLocations(_ context.Context, size int) ([]*Location, error) {
	return d.ranking.Top(size) //@TODO: HERE CONTEXT
}
