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
	ErrInvalidArgument = errors.New("invalid Arguments")
	ErrEmptyCity       = errors.New("bad Request, void City")
)

type SearchedLocationsRanking interface {
	GetTopSearchedLocations(ctx context.Context, size int) ([]*provider.Location, error)
	IncreaseCityScore(ctx context.Context, city string) error
}

type DefaultService struct {
	repository provider.GithubRepository
	ranking    SearchedLocationsRanking
}

func New(r provider.GithubRepository, rnk SearchedLocationsRanking) *DefaultService {
	return &DefaultService{
		repository: r,
		ranking:    rnk,
	}
}

func (s *DefaultService) GetTopContributors(ctx context.Context, r provider.GithubTopRequest) ([]*provider.Contributor, error) {
	log.Infof("GetTopContributors , city: %s size: %d sort %s", r.City, r.Size, r.Sort)
	err := s.ranking.IncreaseCityScore(ctx, r.City)
	if err != nil {
		log.Errorf("unexpected error increasing city score, error %v", err)
	}
	return s.repository.GetGithubTopContributors(ctx, r)
}

func (s *DefaultService) GetTopSearchedLocations(ctx context.Context, size int) ([]*provider.Location, error) {
	log.Infof("GetTopSearchedLocations , size: %d ", size)

	return s.ranking.GetTopSearchedLocations(ctx, size)
}
