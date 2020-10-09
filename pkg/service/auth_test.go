package service

import (
	"context"
	"testing"
	"time"
)

func TestAuthWorkFlowOnValidCredentials(t *testing.T) {
	val := &fakeAuth{valid: true}
	a := NewAuth(val, "../../config/app.rsa", "../../config/app.rsa.pub", time.Minute, "fakeApp")

	token, err := a.Authorize(context.Background(), "test", "known")
	if err != nil {
		t.Fatalf("Unexpected error generating token key, error %s", err.Error())
	}

	valid, err := a.IsValidToken(context.Background(), token)
	if err != nil {
		t.Fatalf("Unexpected error validating token key, error %s", err.Error())
	}

	if !valid {
		t.Error("Expected validated")
	}
}

func TestAuthWorkFlowOnInValidCredentials(t *testing.T) {
	val := &fakeAuth{valid: false}
	a := NewAuth(val, "../../config/app.rsa", "../../config/app.rsa.pub", time.Minute, "fakeApp")

	token, err := a.Authorize(context.Background(), "test", "fooBar")
	if err == nil {
		t.Fatalf("Expected invalid credentials error")
	}

	if err.Error() != ErrUnauthorized.Error() {
		t.Error("UnExpected error message")
	}

	if token != "" {
		t.Error("Expected void token")
	}
}

func TestValidateOnInvalidCredentials(t *testing.T) {
	val := &fakeAuth{}
	a := NewAuth(val, "../../config/app.rsa", "../../config/app.rsa.pub", time.Minute, "fakeApp")
	i, err := a.IsValidToken(context.Background(), "fooBarToken")
	if err == nil {
		t.Fatalf("Expected error validating token key, error %s", err.Error())
	}

	if err.Error() != ErrUnauthorized.Error() {
		t.Error("UnExpected error message")
	}

	if i {
		t.Error("Expected Invalid credentials")
	}
}

type fakeAuth struct {
	valid bool
	err   error
}

func (f *fakeAuth) ValidCredentials(user, pass string) (bool, error) {
	return f.valid, f.err
}
