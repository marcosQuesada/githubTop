package provider

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGithubClientOnFakeServerWithSuccess(t *testing.T) {
	defer func() {
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	var timeout = time.Second * 1

	// Fake Server sleeps to force timeout on the first two request, third one is successful
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "1000")
		w.Header().Set("X-RateLimit-Reset", "1000")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(fakeResponse))
	}))

	token := "fakeToken"
	cfg := HttpConfig{
		OauthToken:      token,
		Timeout:         timeout,
		RateLimitConfig: NewRateLimitConfig(time.Second*1000, 1000),
		Retries:         1000,
	}
	c := NewGithubClient("Test", cfg)

	u, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("Unexpected error, err %s", err.Error())
	}

	//rewrite url to point local server
	c.setURL(u)

	req := GithubTopRequest{
		City:    "barcelona",
		Size:    2,
		Version: APIv1,
	}
	response, err := c.DoRequest(context.Background(), req, 1, 2)
	if err != nil {
		t.Fatalf("Unexpected error, err %s", err.Error())
	}

	if len(response) != 2 {
		t.Fatalf("Unexpected response size, expected 2 got %d", len(response))
	}

	if response[0].Name != "kristianmandrup" {
		t.Error("Unexpected first response")
	}
}

func TestGithubClientOnApiV2WithFakeServerWithSuccess(t *testing.T) {
	defer func() {
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	var timeout = time.Second * 1

	// Fake Server sleeps to force timeout on the first two request, third one is successful
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "1000")
		w.Header().Set("X-RateLimit-Reset", "1000")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(fakeResponse))
	}))

	token := "fakeToken"
	cfg := HttpConfig{
		OauthToken: token,
		Timeout:    timeout,
	}
	c := NewGithubClient("Test", cfg)

	u, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("Unexpected error, err %s", err.Error())
	}

	//rewrite url to point local server
	c.setURL(u)

	req := GithubTopRequest{
		City:    "barcelona",
		Size:    2,
		Version: APIv2,
	}
	response, err := c.DoRequest(context.Background(), req, 1, 2)
	if err != nil {
		t.Fatalf("Unexpected error, err %s", err.Error())
	}

	if len(response) != 2 {
		t.Fatalf("Unexpected response size, expected 2 got %d", len(response))
	}

	if response[0].Email != "foo@bar.com" {
		t.Error("Unexpected email response")
	}

	if response[0].Bio != "fakeBio" {
		t.Error("Unexpected email response")
	}

	if response[0].Company != "fakeCompany" {
		t.Error("Unexpected email response")
	}
}

func TestGithubClientOnFakeServerWithTimeout(t *testing.T) {
	defer func() {
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	var timeout = time.Millisecond * 100

	// Fake Server sleeps to force timeout on the first two request, third one is successful
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to force request timeout
		time.Sleep(time.Millisecond * 200)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fakeResponse))
	}))

	token := "fakeToken"
	cfg := HttpConfig{
		OauthToken: token,
		Timeout:    timeout,
	}
	c := NewGithubClient("Test", cfg)

	u, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Errorf("Unexpected error, err %s", err.Error())
	}

	//rewrite url to point local server
	c.setURL(u)

	req := GithubTopRequest{
		City:    "barcelona",
		Size:    2,
		Version: APIv1,
	}
	_, err = c.DoRequest(context.Background(), req, 1, 2)

	if err != context.DeadlineExceeded {
		t.Error("Unexpected Context timeout error!")
	}
}
