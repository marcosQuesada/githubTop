package service

import (
	"context"
	"errors"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
)

const (
	// available Sizes
	SmallSize  = 50
	MediumSize = 100
	LargeSize  = 150
)

var (
	// ErrInvalidArgument happens on request without valid arguments
	ErrInvalidArgument = errors.New("invalid Arguments")
	// ErrEmptyCity happens on request without defined city
	ErrEmptyCity = errors.New("bad Request, void City")
)

// SearchedLocationsRanking defines location ranking
type SearchedLocationsRanking interface {
	GetTopSearchedLocations(ctx context.Context, size int) ([]*provider.Location, error)
	IncreaseCityScore(ctx context.Context, city string) error
}

// DefaultService defines core service
type DefaultService struct {
	repository provider.GithubRepository
	ranking    SearchedLocationsRanking
}

// New instantiates service
func New(r provider.GithubRepository, rnk SearchedLocationsRanking) *DefaultService {
	return &DefaultService{
		repository: r,
		ranking:    rnk,
	}
}

// GetTopContributors returns github top by location
func (s *DefaultService) GetTopContributors(ctx context.Context, r provider.GithubTopRequest) ([]*provider.Contributor, error) {
	log.Infof("GetTopContributors , city: %s size: %d sort %s", r.City, r.Size, r.Sort)
	err := s.ranking.IncreaseCityScore(ctx, r.City)
	if err != nil {
		log.Errorf("unexpected error increasing city score, error %v", err)
	}
	return s.repository.GetGithubTopContributors(ctx, r)
}

// GetTopSearchedLocations return top Searched Locations
func (s *DefaultService) GetTopSearchedLocations(ctx context.Context, size int) ([]*provider.Location, error) {
	log.Infof("GetTopSearchedLocations , size: %d ", size)

	return s.ranking.GetTopSearchedLocations(ctx, size)
}
