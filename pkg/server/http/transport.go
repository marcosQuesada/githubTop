package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/ratelimit"
	"github.com/google/go-github/github"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"io/ioutil"
	"net/http"
	"strconv"
)

// ErrDefaul happens on unknown source error
var ErrDefaul = errors.New("Unexpected Error.")

// TopContributorsRequest defines api request
type TopContributorsRequest struct {
	City  string
	Size  int
	Token string
}

// TopContributorsResponse defines api response
type TopContributorsResponse struct {
	Top []*provider.Contributor
}

func topContributorsRequestDecoder(_ context.Context, r *http.Request) (interface{}, error) {
	var token = ""

	tokenCookie, err := r.Cookie(service.TokenName)
	// Set token just on success, let errors be tracked by service
	if err == nil {
		token = tokenCookie.Value
	}

	// Let service track all request, even the bad ones, that's why we don't return error here
	city := r.URL.Query().Get("city")
	rawSize := r.URL.Query().Get("size")

	var size = int64(0)
	if rawSize != "" {
		size, err = strconv.ParseInt(rawSize, 10, 0)
		if err != nil {
			log.Errorf("Bad request, error parsing size, err %s", err.Error())
		}
	}

	return TopContributorsRequest{City: city, Size: int(size), Token: token}, nil
}

// AuthRequest defines auth request
type AuthRequest struct {
	User, Pass string
}

// AuthResponse defines auth response
type AuthResponse struct {
	Token string
}

func authDecoder(_ context.Context, r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Body.Close()
	}()

	authResponse := AuthRequest{}
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		log.Errorf("Unexpected error unMarshalling request, err: %s", err.Error())
	}

	return authResponse, err
}

func authResponseEncoder(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	ar, ok := response.(AuthResponse)
	if !ok {
		err := fmt.Errorf("Unexpected response type, %v", response)
		log.Error(err)
		errorEncoder(ctx, err, w)

		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:       service.TokenName,
		Value:      ar.Token,
		Path:       "/",
		RawExpires: "0",
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	return json.NewEncoder(w).Encode(response)
}

func responseEncoder(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	return json.NewEncoder(w).Encode(response)
}

// encode errors from service layer
func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case service.ErrUnauthorized:
		w.WriteHeader(http.StatusForbidden)

	case service.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)

	case ratelimit.ErrLimited:
		err = fmt.Errorf("API rate limit exceeded")
		w.WriteHeader(http.StatusTooManyRequests)

	case context.DeadlineExceeded:
		w.WriteHeader(http.StatusRequestTimeout)

	default:
		switch e := err.(type) {
		case *github.RateLimitError:
			err = fmt.Errorf("API rate limit exceeded, reset on: %s", e.Rate.Reset.String())
			w.WriteHeader(http.StatusTooManyRequests)

		case *github.AbuseRateLimitError:
			err = fmt.Errorf("abuse Rate Limit error, retry after %s", e.RetryAfter.String())
			w.WriteHeader(http.StatusTooManyRequests)

		case *github.ErrorResponse:
			w.WriteHeader(http.StatusServiceUnavailable)

		default:
			err = ErrDefaul
			w.WriteHeader(http.StatusInternalServerError)
		}

	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
