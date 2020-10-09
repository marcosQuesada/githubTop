package service

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"testing"
)

func TestDefaultContributorServiceOnFakeRepositoryOnSinglePage(t *testing.T) {
	r := newFakeRepository(50)
	s := New(r, "fakeApp")

	tc, err := s.GetTopContributors(context.Background(), "foo", 50)
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
	s := New(r, "fakeApp")

	tc, err := s.GetTopContributors(context.Background(), "foo", 150)
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

func (f *fakeRepository) GetGithubTopContributors(ctx context.Context, city string, size int) ([]*provider.Contributor, error) {
	return f.items, nil
}
