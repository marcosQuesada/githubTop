package provider

import (
	"context"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/upgear/go-kit/retry"
	"sync"
	"time"
)

const (
	MaxPerPage = 100
)

// Contributor models github contributor
type Contributor struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Url     string `json:"url"`
	Company string `json:"company,omitempty"`
	Email   string `json:"email,omitempty"`
	Bio     string `json:"bio,omitempty"`
}

// GithubRepository defines github repository
type GithubRepository interface {
	GetGithubTopContributors(ctx context.Context, req GithubTopRequest) ([]*Contributor, error)
}

type httpGithubRepository struct {
	client  *GithubClient
	retries int
}

// GithubResult defines github contributors top query result
type GithubTopRequest struct {
	City    string
	Size    int
	Version string
	Sort    string
}

// GithubResult defines github contributors top query result
type GithubResult struct {
	res  []*Contributor
	page int
	err  error
}

// HttpConfig instantiates an http config
type HttpConfig struct {
	OauthToken      string
	Timeout         time.Duration
	Retries         int
	RateLimitConfig RateLimitConfig
}

// RateLimitConfig defines rate limiter config
type RateLimitConfig struct {
	RateLimitWindow time.Duration
	RateLimitMaxReq int
}

// NewRateLimitConfig creates rate limiter config
func NewRateLimitConfig(w time.Duration, l int) RateLimitConfig {
	return RateLimitConfig{w, l}
}

// NewHttpGithubRepository instantiates http github repository
func NewHttpGithubRepository(appName string, cfg HttpConfig) *httpGithubRepository {
	c := NewGithubClient(appName, cfg)

	return &httpGithubRepository{
		client:  c,
		retries: cfg.Retries,
	}
}

// GetGithubTopContributors gets github top contributors
func (r *httpGithubRepository) GetGithubTopContributors(ctx context.Context, req GithubTopRequest) ([]*Contributor, error) {
	var response []*Contributor

	// apply retry policy
	err := retry.Double(r.retries).Run(func() error {
		var err error

		response, err = r.getGithubTopContributors(context.Background(), req)
		if err != nil {
			// on rate limit errors stop retry
			if !retryOnResponseError(err) {
				_ = retry.Stop(err)
			}
		}

		return err
	})

	return response, err
}

func (r *httpGithubRepository) getGithubTopContributors(ctx context.Context, req GithubTopRequest) ([]*Contributor, error) {
	rp := paginateRequest(req.Size)

	wg := sync.WaitGroup{}
	wg.Add(len(rp))
	res := make(chan GithubResult, len(rp))

	// Run page requests concurrently
	for _, v := range rp {
		go func(vv requestPage) {
			tcp, err := r.client.DoRequest(ctx, req, vv.page, vv.size)
			res <- GithubResult{res: tcp, err: err, page: vv.page}
			wg.Done()
		}(v)
	}

	//wait results
	wg.Wait()
	close(res)

	//ensure order!
	ogr := make([]GithubResult, len(rp))
	for g := range res {
		if g.err != nil {
			log.Errorf("Unexpected error reading results, err %s", g.err.Error())
			return nil, g.err
		}

		ogr[g.page-1] = g
	}

	//concatenate results
	cb := make([]*Contributor, 0)
	for _, g := range ogr {

		cb = append(cb, g.res...)
	}

	log.Infof("TOTAL ENTRIES %d", len(cb))
	return cb, nil

}

type requestPage struct {
	page, size int
}

func paginateRequest(size int) []requestPage {
	res := []requestPage{}
	pages := int((size-1)/100) + 1
	for i := 1; i <= pages; i++ {
		var itemsPerPage = size
		if size > MaxPerPage {
			itemsPerPage = MaxPerPage
		}

		size = size - itemsPerPage

		res = append(res, requestPage{i, itemsPerPage})
	}

	return res
}
