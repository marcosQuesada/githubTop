package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var dataProvider = []struct {
	uri          string
	expectedSize int
	city         string
	err          error
}{
	{"", 0, "", service.ErrEmptyCity},
	{"?size=150", 150, "", service.ErrEmptyCity},
	{"?city=barcelona&size=150", 150, "barcelona", nil},
	{"?size=aaaaa", 0, "", service.ErrEmptyCity},
}

func TestDecodeTopContributorsRequestCornerCases(t *testing.T) {
	for _, v := range dataProvider {
		uri := fmt.Sprintf("http://localhost:8000/top/v1%s", v.uri)
		req := httptest.NewRequest("GET", uri, nil)
		d, err := topContributorsRequestDecoder(context.Background(), req)
		if err != v.err {
			t.Errorf("Unexpected error decoding request, err %v", err)
		}

		if err != nil {
			continue
		}

		tcr, ok := d.(TopContributorsRequest)
		if !ok {
			t.Fatalf("Unexpected request type, has %T", d)
		}

		if tcr.Size != v.expectedSize {
			t.Errorf("Unexpected size, expected %d but got %d", v.expectedSize, tcr.Size)
		}

		if tcr.City != v.city {
			t.Errorf("Unexpected city, expected %s but got %s", v.city, tcr.City)
		}
	}
}

func TestAuthDecoderRequest(t *testing.T) {

	cr := map[string]string{"user": "test", "pass": "known"}
	rawCr, err := json.Marshal(cr)
	if err != nil {
		t.Fatalf("Unexpected error marsalling data, err %s", err.Error())
	}
	req, err := http.NewRequest("POST", "http://localhost:8000/auth", bytes.NewBuffer(rawCr))
	if err != nil {
		t.Fatalf("Unexpected error creating request, err %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	d, err := authDecoder(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error decoding auth request, err %s", err.Error())
	}

	r, ok := d.(AuthRequest)
	if !ok {
		t.Errorf("Unexpected type, got %t", d)
	}

	if r.User != "test" {
		t.Errorf("Unexpected user value, got %s", r.User)
	}

	if r.Pass != "known" {
		t.Errorf("Unexpected password value, got %s", r.Pass)
	}
}

func TestAuthResponseEncoder(t *testing.T) {
	rec := NewFakeResponseRecorder()
	ar := AuthResponse{Token: "fakeToken"}
	err := authResponseEncoder(context.Background(), rec, ar)
	if err != nil {
		t.Errorf("Unexpected error on auth response encoder, err %s", err.Error())
	}

	rawCookie := rec.HeaderMap.Get("Set-Cookie")
	parts := strings.Split(rawCookie, ";")
	parts = strings.Split(parts[0], "=")

	if parts[1] != "fakeToken" {
		t.Errorf("Unexpected raw cookie response %s", rawCookie)
	}
}

type fakeResponseRecorder struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to
	Flushed   bool
}

// NewFakeResponseRecorder returns an initialized fakeResponseRecorder.
func NewFakeResponseRecorder() *fakeResponseRecorder {
	return &fakeResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
	}
}

// Header returns the response headers.
func (rw *fakeResponseRecorder) Header() http.Header {
	return rw.HeaderMap
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *fakeResponseRecorder) Write(buf []byte) (int, error) {
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	if rw.Code == 0 {
		rw.Code = http.StatusOK
	}
	return len(buf), nil
}

// WriteHeader sets rw.Code.
func (rw *fakeResponseRecorder) WriteHeader(code int) {
	rw.Code = code
}

// Flush sets rw.Flushed to true.
func (rw *fakeResponseRecorder) Flush() {
	rw.Flushed = true
}
