package cmd

import (
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	httpServer "github.com/marcosQuesada/githubTop/pkg/server/http"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	APP_NAME = "GithubTop"
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
		repo := provider.NewHttpGithubRepository(APP_NAME, cfg)
		cache := provider.NewGithubRepositoryCache(cacheCfg, repo)

		svc := service.New(cache, APP_NAME)
		ac := service.NewDefaultStaticAuthorizer()
		auth := service.NewAuth(ac, "config/app.rsa", "config/app.rsa.pub", tokenTTL, APP_NAME)
		s := httpServer.New(port, svc, auth, APP_NAME)

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

		err := s.Run()
		if err != nil {
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
