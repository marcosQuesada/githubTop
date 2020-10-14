package provider

import (
	"context"

	"testing"


)

func TestRepositoryMiddlewareOnCacheHit(t *testing.T) {
	c := []*Contributor{{ID: 1}, {ID: 2}}
	ch := &fakeCache{contributors: c}
	repo := &fakeRepository{}
	r := NewCacheMiddleware(ch, repo)

	err := r.AddTopContributors("barcelona", 2, c)

	if err != nil {
		t.Errorf("Unexpected error adding contributors, err: %s", err.Error())

	}

	expected := 1
	if ch.called != expected {
		t.Fatalf("unexpected cache total calls, expected %d got %d", expected, ch.called)
	}

	v, err := r.GetGithubTopContributors(context.Background(), "barcelona", 2)
	if err != nil {
		t.Errorf("Unexpected error getting contributors, err: %s", err.Error())
	}

	if len(v) != 2 {
		t.Errorf("Unexpected contributors size, expected 2 got %d", len(v))
	}

	if ch.called != expected {
		t.Fatalf("unexpected cache total calls, expected %d got %d", expected, ch.called)
	}

	expected = 0
	if repo.called != expected {
		t.Fatalf("unexpected repository total calls, expected %d got %d", expected, ch.called)
	}
}

func TestRepositoryMiddlewareOnCacheMiss(t *testing.T) {
	c := []*Contributor{{ID: 1}, {ID: 2}}
	ch := &fakeCache{}
	repo := &fakeRepository{contributors: c}
	r := NewCacheMiddleware(ch, repo)

	err := r.AddTopContributors("barcelona", 2, c)

	if err != nil {
		t.Errorf("Unexpected error adding contributors, err: %s", err.Error())

	}

	expected := 1

	if ch.called != expected {
		t.Fatalf("unexpected cache total calls, expected %d got %d", expected, ch.called)

	}

	v, err := r.GetGithubTopContributors(context.Background(), "barcelona", 2)
	if err != nil {
		t.Errorf("Unexpected error getting contributors, err: %s", err.Error())
	}

	if len(v) != 2 {
		t.Errorf("Unexpected contributors size, expected 2 got %d", len(v))
	}
	expected = 2
	if ch.called != expected {
		t.Fatalf("unexpected cache total calls, expected %d got %d", expected, ch.called)
	}

	expected = 1
	if repo.called != expected {
		t.Fatalf("unexpected repository total calls, expected %d got %d", expected, ch.called)
	}
}

type fakeRepository struct {
	contributors []*Contributor
	called       int
}

func (f *fakeRepository) GetGithubTopContributors(ctx context.Context, city string, size int) ([]*Contributor, error) {
	f.called++
	return f.contributors, nil
}

type fakeCache struct {
	contributors []*Contributor
	called       int
}

func (f *fakeCache) Add(k string, v interface{}) error {
	f.called++
	return nil
}

func (f *fakeCache) Get(k string) (interface{}, error) {
	if len(f.contributors) == 0 {
		return nil, ErrCacheMiss
	}
	return f.contributors, nil
}

func (f *fakeCache) Terminate() {}
