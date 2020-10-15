package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/google/go-github/github"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/metrics"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
	"net/url"
	"time"
)

const (
	SortByRepositories = "repositories"
	SortByCommits      = "commits"
	SortByLabels       = "labels"
	Query              = "location:%s"
	APIv1              = "v1"
	APIv2              = "v2"
)

// ErrMaxRetries happens on max request achieved
var ErrMaxRetries = errors.New("max retries achieved")

// GithubClient builds a github http client
type GithubClient struct {
	timeout  time.Duration
	endpoint endpoint.Endpoint

	//useful to allow proper testing, as we need to rewrite BaseURL to point fake Server
	client *github.Client
}

// GithubRequest defines request
type GithubRequest struct {
	Query   string
	Options *github.SearchOptions
}

// GithubResponse defines response
type GithubResponse struct {
	Result   *github.UsersSearchResult
	Response *github.Response
}

// NewGithubClient instantiates github http client
func NewGithubClient(appName string, cfg HttpConfig) *GithubClient {
	client := buildGithubClient(cfg.OauthToken)

	// Wrap Endpoint with rate limiter
	e := makeGithubClientEndpoint(client)
	e = metrics.InstrumentingMiddleware(appName, "githubClient")(e)
	e = log.LoggingMiddleware(kitlog.With(log.MiddlewareLogger, "method", "githubClient"))(e)
	limit := rate.NewLimiter(rate.Every(cfg.RateLimitConfig.RateLimitWindow), cfg.RateLimitConfig.RateLimitMaxReq)
	e = ratelimit.NewErroringLimiter(limit)(e)

	return &GithubClient{
		endpoint: e,
		timeout:  cfg.Timeout,
		client:   client,
	}
}

// DoRequest fires http request
func (r *GithubClient) DoRequest(ctx context.Context, req GithubTopRequest, page, size int) ([]*Contributor, error) {
	query := fmt.Sprintf(Query, req.City)
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: size,
		},
		Sort: req.Sort,
	}

	//Add execution timeout deadline
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.endpoint(ctx, GithubRequest{query, opt})
	if err != nil {
		log.Errorf("Endpoint error %v", err)
		return nil, err
	}

	response := res.(GithubResponse)
	log.Infof("Remaining is %d", response.Response.Remaining)
	if response.Response.Remaining == 0 {
		log.Error("Max Request by Minute achieved")
		return nil, ErrMaxRetries
	}

	cs := make([]*Contributor, size)
	for k, u := range response.Result.Users {
		c := &Contributor{ID: *u.ID, Name: *u.Login, Url: *u.URL}
		if req.Version == APIv2 {
			if u.Company != nil {
				c.Company = *u.Company
			}
			if u.Email != nil {
				c.Email = *u.Email
			}
			if u.Bio != nil {
				c.Bio = *u.Bio
			}
		}
		cs[k] = c
	}
	return cs, nil
}

// setURL replaces github baseURL, useful on testing
func (r *GithubClient) setURL(u *url.URL) {
	r.client.BaseURL = u
}

func makeGithubClientEndpoint(client *github.Client) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(GithubRequest)
		res, resp, err := client.Search.Users(ctx, req.Query, req.Options)

		return GithubResponse{res, resp}, err
	}
}

func retryOnResponseError(err error) (retry bool) {
	switch err.(type) {
	case *github.AcceptedError:
		return true
	case *github.RateLimitError:
		return false
	case *github.AbuseRateLimitError:
		return false

	case error:
		if err == context.DeadlineExceeded {
			return true

		} else if err == ratelimit.ErrLimited {
			return false
		}

		log.Errorf("Unexpected error searching github users, err %s", err.Error())

		return true
	default:
		return true
	}
}

func buildGithubClient(oauthToken string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauthToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
