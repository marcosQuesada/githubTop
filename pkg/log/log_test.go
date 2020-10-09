package log

import (
	"context"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	l := &fakeLogger{}
	m := LoggingMiddleware(l)

	e := &fakeEndpoint{}
	r := m(e.endpoint)

	p, err := r(context.Background(), struct{}{})
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}

	if p != 1 {
		t.Errorf("Unexpected result, expected 1 has %d", p)
	}

	if !e.called {
		t.Error("Endpoint has not been called")
	}

	if !l.called {
		t.Error("Instrumented Logger has not been called")
	}
}

type fakeEndpoint struct {
	called bool
}

func (f *fakeEndpoint) endpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	f.called = true
	return 1, nil
}

type fakeLogger struct {
	called bool
}

func (f *fakeLogger) Log(keyvals ...interface{}) error {
	f.called = true
	Info(keyvals...)
	return nil
}
