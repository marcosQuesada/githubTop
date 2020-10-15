package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"testing"
)

func TestDefaultContributorServiceOnFakeRepositoryOnSinglePage(t *testing.T) {
	r := newFakeRepository(50)
	rnk := &fakeRanking{}
	s := New(r, rnk)
	req := provider.GithubTopRequest{
		City:    "foo",
		Size:    50,
		Version: provider.APIv1,
	}
	tc, err := s.GetTopContributors(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error getting contributors, err %s", err.Error())
	}

	if len(tc) != 50 {
		t.Errorf("Unexpected result size, expected 150 got %d", len(tc))
	}

	if tc[49].Name != "fakeUser_49" {
		t.Errorf("Unexpected last contributor name, got %s", tc[149].Name)
	}
}

func TestDefaultContributorServiceOnFakeRepository(t *testing.T) {
	r := newFakeRepository(150)
	rnk := &fakeRanking{}
	s := New(r, rnk)

	req := provider.GithubTopRequest{
		City:    "barcelona",
		Size:    150,
		Version: provider.APIv1,
	}
	tc, err := s.GetTopContributors(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error getting contributors, err %s", err.Error())
	}

	if len(tc) != 150 {
		t.Errorf("Unexpected result size, expected 150 got %d", len(tc))
	}

	if tc[149].Name != "fakeUser_149" {
		t.Errorf("Unexpected last contributor name, got %s", tc[149].Name)
	}
}

type fakeRepository struct {
	items []*provider.Contributor
}

func newFakeRepository(totalItems int) *fakeRepository {

	items := make([]*provider.Contributor, totalItems)

	for id := 0; id < totalItems; id++ {

		items[id] = &provider.Contributor{ID: int64(id), Name: fmt.Sprintf("fakeUser_%d", id)}
	}

	return &fakeRepository{
		items: items,
	}
}

func (f *fakeRepository) GetGithubTopContributors(ctx context.Context, req provider.GithubTopRequest) ([]*provider.Contributor, error) {
	return f.items, nil
}

type fakeRanking struct{}

func (f *fakeRanking) GetTopSearchedLocations(ctx context.Context, size int) ([]*provider.Location, error) {
	return []*provider.Location{
		{Name: "barcelona", Score: 1000}, {Name: "badalona", Score: 10},
	}, nil
}

func (f *fakeRanking) IncreaseCityScore(ctx context.Context, city string) error {
	return nil
}
