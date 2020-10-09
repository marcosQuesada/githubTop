package metrics

import (
	"context"
	"github.com/go-kit/kit/metrics"
	"testing"
)

func TestInstrumentingMiddleware(t *testing.T) {
	success := &fakeCounter{}
	fail := &fakeCounter{}
	duration := &fakeHistogram{}

	i := instrumentingMiddlewareWithMetrics(success, fail, duration)

	e := &fakeEndpoint{}
	r := i(e.endpoint)

	p, err := r(context.Background(), struct{}{})
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}

	if p != 1 {
		t.Errorf("Unexpected result, expected 1 has %d", p)
	}

	if !success.called {
		t.Error("Unexpected result, counter has not been called")
	}

	if duration.called != 1 {
		t.Errorf("Unexpected Histogram called times, expected 1 has %d", duration.called)
	}
}

type fakeEndpoint struct {
	called bool
}

func (f *fakeEndpoint) endpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	f.called = true
	return 1, nil
}

type fakeCounter struct {
	called bool
}

func (f *fakeCounter) With(labelValues ...string) metrics.Counter {
	return f
}

func (f *fakeCounter) Add(delta float64) {
	f.called = true
}

type fakeHistogram struct {
	called int
}

func (f *fakeHistogram) With(labelValues ...string) metrics.Histogram {
	f.called++
	return f
}
func (f *fakeHistogram) Observe(value float64) {

}
