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

type Service interface {
	GetTopContributors(ctx context.Context, city string, size int) ([]*provider.Contributor, error)
}

type DefaultService struct {
	repository provider.GithubRepository
}

func New(r provider.GithubRepository) *DefaultService {
	return &DefaultService{
		repository: r,
	}
}

func (s *DefaultService) GetTopContributors(ctx context.Context, city string, size int) ([]*provider.Contributor, error) {
	log.Infof("GetTopContributors , city: %s size: %d ", city, size)
	return s.repository.GetGithubTopContributors(ctx, city, size)
}
