package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestEndpointDevelopmentFlow(t *testing.T) {
	s := &Server{}

	svc := &fakeService{}
	h := s.makeTopContributorsHandler(svc, "fakeApp")
	svr := httptest.NewServer(h)

	defer func() {
		svr.Close()
		// Clean metrics registry on test done
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	expectedSize := 150

	var netClient = &http.Client{}
	resp, err := netClient.Get(fmt.Sprintf("%s?city=barcelona&size=%d", svr.URL, expectedSize))
	if err != nil {
		log.Fatalf("Response error %v", err)
	}

	if svc.getRequestedSize() != expectedSize {
		t.Error("Unexpected size")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected %v error reading body content ", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	e := make(map[string][]*provider.Contributor)
	err = json.Unmarshal(body, &e)
	if err != nil {
		log.Fatalf("unexpected %v error unmarshalling body content ", err)
	}
	v, ok := e["Top"]
	if !ok {
		t.Error("Unexpected error, top field not exists on response")
	}

	if len(v) != 1 {
		t.Fatalf("Unexpected response size, expected 1 got %d", len(v))
	}

	if v[0].Name != "foo" {
		t.Errorf("Unexpected contributors response name, got %s", v[0].Name)
	}
}

func TestOnEndpointErrorResponseHasErrorEncodedAsResponseBody(t *testing.T) {
	s := &Server{}
	expectedError := errors.New("Fake Error.")
	svc := &fakeService{err: expectedError}
	h := s.makeTopContributorsHandler(svc, "fakeApp")
	svr := httptest.NewServer(h)

	defer func() {
		svr.Close()
		// Clean metrics registry on test done
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	expectedSize := 150

	netClient := &http.Client{}
	resp, err := netClient.Get(fmt.Sprintf("%s?city=barcelona&size=%d", svr.URL, expectedSize))
	if err != nil {
		t.Fatalf("Unexpected response error, err %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Unexpected status code, expected %d but got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error reading body, err %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	e := make(map[string]string)
	err = json.Unmarshal(body, &e)
	if err != nil {
		t.Fatalf("Unexpected error deserializing body, err %s", err.Error())
	}

	errorMessage, ok := e["error"]
	if !ok {
		t.Fatal("Expected error field not found!")
	}

	if errorMessage != ErrUnexpected.Error() {
		t.Errorf("Unexpected error message, expected %s but got %s", expectedError.Error(), errorMessage)
	}
}

func TestAuthEndpointWorkFlowOnValidCredentials(t *testing.T) {
	s := &Server{}
	svc := &fakeAuthService{}
	h := s.makeAuthTransport(svc, "fakeApp")
	svr := httptest.NewServer(h)

	defer func() {
		svr.Close()
		// Clean metrics registry on test done
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	cr := map[string]string{"user": "test", "pass": "known"}
	rawCr, err := json.Marshal(cr)
	if err != nil {
		t.Fatalf("Unexpected error marsalling data, err %s", err.Error())
	}
	req, err := http.NewRequest("POST", svr.URL, bytes.NewBuffer(rawCr))
	if err != nil {
		t.Fatalf("Unexpected request creation error, err %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unexpected response error, err %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code, expected %d but got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error reading body, err %s", err.Error())
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data := map[string]string{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatalf("Unexpected error unmarshalling body, err %s", err.Error())
	}

	v, ok := data["Token"]
	if !ok {
		t.Error("Unexpected response, token field not found")
	}

	if v != "fakeToken" {
		t.Errorf("Unexpected FakeToken response, got %s", v)
	}
}

func TestMakeAuthTopContributorsHandlerOnFakeAuthServicePassingCredentialsDoesNormalFlow(t *testing.T) {
	s := &Server{}
	svc := &fakeService{}
	authSvc := &fakeAuthService{"fakeToken"}

	h := s.makeAuthTopContributorsHandler(svc, authSvc, "fakeApp")
	svr := httptest.NewServer(h)

	defer func() {
		svr.Close()
		// Clean metrics registry on test done
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
	}()

	var netClient = &http.Client{}

	expectedSize := 150

	uri := fmt.Sprintf("%s?city=barcelona&size=%d", svr.URL, expectedSize)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating request, err %s", err.Error())
	}

	cookie := http.Cookie{Name: "AccessToken", Value: "fakeToken"}
	req.AddCookie(&cookie)

	resp, err := netClient.Do(req)
	if err != nil {
		t.Fatalf("Unexpected error executing request, err %s", err.Error())
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if svc.getRequestedSize() != expectedSize {
		t.Error("Unexpected size")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unexpected error readyng response body, err %s", err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	e := make(map[string][]*provider.Contributor)
	err = json.Unmarshal(body, &e)
	if err != nil {
		t.Fatalf("Unexpected error unmarshalling body, err %s", err.Error())
	}

	v, ok := e["Top"]
	if !ok {
		t.Error("Unexpected error, top field not exists on response")
	}

	if len(v) != 1 {
		t.Fatalf("Unexpected response size, expected 1 got %d", len(v))
	}

	if v[0].Name != "foo" {
		t.Errorf("Unexpected contributors response name, got %s", v[0].Name)
	}
}

type fakeService struct {
	requestSize int
	mutex       sync.RWMutex
	err         error
}

// GetTopContributors fake method
func (s *fakeService) GetTopContributors(_ context.Context, r provider.GithubTopRequest) ([]*provider.Contributor, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.requestSize = r.Size

	return []*provider.Contributor{{ID: 1, Name: "foo"}}, s.err
}

func (s *fakeService) getRequestedSize() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.requestSize
}

func (s *fakeService) GetTopSearchedLocations(_ context.Context, size int)([]*provider.Location, error) {
	return []*provider.Location{
		{Name: "barcelona", Score: 1000},{Name: "badalona", Score: 10},
	}, nil
}

type fakeAuthService struct {
	token string
}

// Authorize fake method
func (f *fakeAuthService) Authorize(_ context.Context, user, pass string) (token string, err error) {
	return "fakeToken", nil
}

// IsValidToken fake method
func (f *fakeAuthService) IsValidToken(_ context.Context, token string) (bool, error) {
	return f.token == token, nil
}
