package cmd

import (
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"github.com/marcosQuesada/githubTop/pkg/provider/cache"
	httpServer "github.com/marcosQuesada/githubTop/pkg/server/http"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	AppName = "GithubTop"
)

var (
	port                 int
	oauthToken           string
	requestTimeout       time.Duration
	requestRetries       int
	cacheTTL             time.Duration
	cacheExpirationFreq  time.Duration
	tokenTTL             time.Duration
	rateLimitWindow      time.Duration
	rateLimitMaxRequests int
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start http server",
	Long:  `Start http server`,
	Run: func(cmd *cobra.Command, args []string) {

		rateCfg := provider.NewRateLimitConfig(rateLimitWindow, rateLimitMaxRequests)
		cfg := provider.HttpConfig{
			OauthToken:      oauthToken,
			Timeout:         requestTimeout,
			Retries:         requestRetries,
			RateLimitConfig: rateCfg,
		}

		cacheCfg := provider.NewCacheConfig(cacheTTL, cacheExpirationFreq)
		repo := provider.NewHttpGithubRepository(AppName, cfg)
		cachePersistence, err := cache.NewLRUCache(cacheCfg.Ttl, cacheCfg.ExpirationFrequency)
		if err != nil {
			log.Fatalf("unexepcted error initializing lru cache, error %v", err)
		}
		cache := provider.NewCacheMiddleware(cachePersistence, repo)

		svc := service.New(cache)
		ac := service.NewDefaultStaticAuthorizer()
		auth := service.NewAuth(ac, "config/app.rsa", "config/app.rsa.pub", tokenTTL, AppName)
		s := httpServer.New(port, svc, auth, AppName)

		c := make(chan os.Signal, 1)

		signal.Notify(
			c,
			os.Interrupt,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		//serve until signal
		go func() {
			<-c
			s.Terminate()
		}()

		if err := s.Run(); err != nil {
			log.Errorf("Unexpected error: %v", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().IntVarP(&port, "port", "p", 8000, "Http Server Port")
	httpCmd.Flags().StringVarP(&oauthToken, "oauth", "0", "", "Github personal Oauth token")
	httpCmd.Flags().DurationVarP(&requestTimeout, "timeout", "t", time.Second*3, "http request timeout")
	httpCmd.Flags().IntVarP(&requestRetries, "retries", "r", 3, "http request on error retry")
	httpCmd.Flags().DurationVarP(&cacheTTL, "cache-ttl", "c", time.Hour*24, "cache TTL")
	httpCmd.Flags().DurationVarP(&cacheExpirationFreq, "cache-exp-freq", "e", time.Second*5, "cache expiration frequency")
	httpCmd.Flags().DurationVarP(&tokenTTL, "token-ttl", "l", time.Minute*1, "auth token expiration")
	httpCmd.Flags().DurationVarP(&rateLimitWindow, "rate-window", "w", time.Minute*1, "rate limit time window")
	httpCmd.Flags().IntVarP(&rateLimitMaxRequests, "rate-max", "m", 30, "rate limit max requests")
}
