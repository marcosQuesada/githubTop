package http

import (
	"context"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"testing"
)

func TestAuthMiddlewareOnValidCredentialsForwardRequestToEndpoint(t *testing.T) {
	svc := &fakeAuthService{"fakeToken"}
	a := authMiddleware(svc)

	e := &fakeEndpoint{}
	r := a(e.endpoint)

	req := TopContributorsRequest{Token: "fakeToken"}
	p, err := r(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}

	// as access is enabled, response has been forwarded to endpoint
	if !e.called {
		t.Errorf("Endpoint has not been called")
	}

	if p != 1 {
		t.Errorf("Unexpected endpoint response, response %d", p)
	}

}

func TestAuthMiddlewareOnInvalidCredentialsDoesNotForwardRequestToEndpoint(t *testing.T) {
	svc := &fakeAuthService{"fakeToken"}
	a := authMiddleware(svc)

	e := &fakeEndpoint{}
	r := a(e.endpoint)

	req := TopContributorsRequest{Token: "invalidToken"}
	_, err := r(context.Background(), req)
	if err == nil {
		t.Error("Expected Forbiden access error ")
	}

	if err.Error() != service.ErrUnauthorized.Error() {
		t.Error("UnExpected error message")
	}

	//unauthorized credentials, does not forward request to fakeEndpoint
	if e.called {
		t.Errorf("Endpoint has been called")
	}
}

type fakeEndpoint struct {
	called bool
}

func (f *fakeEndpoint) endpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	f.called = true
	return 1, nil
}
