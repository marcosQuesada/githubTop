package ranking

import (
	"testing"
)

func TestInMemoryRankingPopulatesRequestedLocations(t *testing.T) {
	maxSize := 5
	r := NewInMemory(maxSize)
	for _, d := range dataProvider {
		for i := 0; i < d.score; i++ {
			err := r.IncreaseScore(d.city)
			if err != nil {
				t.Fatalf("unexpected error populating ranking, error %v", err)
			}
		}
	}

	size := r.Len()

	if size != maxSize {
		t.Fatalf("expected sizes do not match, expcted %d got %d", maxSize, size)
	}
}

func TestInMemoryRankingTopLocations(t *testing.T) {
	maxSize := 10
	r := NewInMemory(maxSize)
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
