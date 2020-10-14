package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestGithubRepositoryOnFakeServerWithRetriesOnTimeoutFiresMaxRetries(t *testing.T) {
	var iterations = 0
	var maxRetries = 3
	var timeout = time.Millisecond * 100
	mutex := sync.RWMutex{}

	// Fake Server sleeps to force timeout on the first two request, third one is successful
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		iterations++
		if iterations < maxRetries {
			// Sleep to force request timeout
			time.Sleep(time.Millisecond * 200)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fakeResponse))
	}))

	token := "fakeToken"
	cfg := HttpConfig{
		OauthToken: token,
		Timeout:    timeout,
		Retries:    maxRetries,
	}
	r := NewHttpGithubRepository("test", cfg)
	u, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("Unexpected error, err %v", err)
	}

	//rewrite url to point local server
	r.client.setURL(u)
	req := GithubTopRequest{
		City:    "barcelona",
		Size:    2,
		Version: APIv1,
	}
	_, err = r.GetGithubTopContributors(context.Background(), req)
	if !errors.Is(err, ErrMaxRetries) {
		t.Errorf("unexpected error getting top contributors, error %v", err)
	}
}

var dataProvider = []struct {
	size     int
	expected []requestPage
}{
	{50, []requestPage{{1, 50}}},
	{100, []requestPage{{1, 100}}},
	{150, []requestPage{{1, 100}, {2, 50}}},
	//check that works fine on larger request
	{450, []requestPage{{1, 100}, {2, 100}, {3, 100}, {4, 100}, {5, 50}}},
}

func TestPaginatorRequest(t *testing.T) {
	for _, v := range dataProvider {
		rp := paginateRequest(v.size)

		if len(rp) != len(v.expected) {
			t.Errorf("Unexpected lenght on paginate, expected %d got %d", len(v.expected), len(rp))
			continue
		}

		for k, r := range rp {
			if r.page != v.expected[k].page {
				t.Errorf("Unexpected request page, expected %d got %d", v.expected[k].page, r.page)
			}

			if r.size != v.expected[k].size {
				t.Errorf("Unexpected request size, expected %d got %d", v.expected[k].size, r.size)

			}
		}

	}
}

var fakeResponse = `{
  "total_count": 9852,
  "incomplete_results": false,
  "items": [
    {
      "login": "kristianmandrup",
      "id": 125005,
      "avatar_url": "https://avatars3.githubusercontent.com/u/125005?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/kristianmandrup",
      "email": "foo@bar.com",
      "company": "fakeCompany",
      "bio": "fakeBio",
      "html_url": "https://github.com/kristianmandrup",
      "followers_url": "https://api.github.com/users/kristianmandrup/followers",
      "following_url": "https://api.github.com/users/kristianmandrup/following{/other_user}",
      "gists_url": "https://api.github.com/users/kristianmandrup/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/kristianmandrup/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/kristianmandrup/subscriptions",
      "organizations_url": "https://api.github.com/users/kristianmandrup/orgs",
      "repos_url": "https://api.github.com/users/kristianmandrup/repos",
      "events_url": "https://api.github.com/users/kristianmandrup/events{/privacy}",
      "received_events_url": "https://api.github.com/users/kristianmandrup/received_events",
      "type": "User",
      "site_admin": false,
      "score": 1.0
    },
    {
      "login": "leobcn",
      "id": 13684313,
      "avatar_url": "https://avatars3.githubusercontent.com/u/13684313?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/leobcn",
	  "email": "foo@bar.com",
      "company": "fakeCompany",
      "bio": "fakeBio",
      "html_url": "https://github.com/leobcn",
      "followers_url": "https://api.github.com/users/leobcn/followers",
      "following_url": "https://api.github.com/users/leobcn/following{/other_user}",
      "gists_url": "https://api.github.com/users/leobcn/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/leobcn/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/leobcn/subscriptions",
      "organizations_url": "https://api.github.com/users/leobcn/orgs",
      "repos_url": "https://api.github.com/users/leobcn/repos",
      "events_url": "https://api.github.com/users/leobcn/events{/privacy}",
      "received_events_url": "https://api.github.com/users/leobcn/received_events",
      "type": "User",
      "site_admin": false,
      "score": 1.0
    }
    ]
}`
