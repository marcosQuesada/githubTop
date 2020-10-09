package service

import (
	"context"
	"errors"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
)

const (
	// available Sizes
	SMALL_SIZE   = 50
	MEDIAUM_SIZE = 100
	LARGE_SIZE   = 150
)

var (
	ErrInvalidArgument = errors.New("Invalid Arguments")
	ErrEmptyCity       = errors.New("Bad Request, void City")
)

type Service interface {
	GetTopContributors(ctx context.Context, city string, size int) ([]*provider.Contributor, error)
}

type DefaultService struct {
	repository provider.GithubRepository
	appName    string
}

func New(c provider.GithubRepository, n string) *DefaultService {
	return &DefaultService{
		repository: c,
		appName:    n,
	}
}

func (s *DefaultService) GetTopContributors(ctx context.Context, city string, size int) ([]*provider.Contributor, error) {
	log.Infof("GetTopContributors , city: %s size: %d ", city, size)
	if city == "" {
		log.Error(ErrEmptyCity)

		return nil, ErrInvalidArgument
	}

	if size != SMALL_SIZE && size != MEDIAUM_SIZE && size != LARGE_SIZE {
		log.Errorf("Bad Request, size %d not permitted", size)

		return nil, ErrInvalidArgument
	}

	v, err := s.repository.GetGithubTopContributors(ctx, city, size)
	if err != nil {
		return nil, err
	}

	return v, nil
}
