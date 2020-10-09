package http

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/metrics"
	"github.com/marcosQuesada/githubTop/pkg/service"
)

func buildMiddleware(n, s string, e endpoint.Endpoint) endpoint.Endpoint {
	e = metrics.InstrumentingMiddleware(n, s)(e)
	e = log.LoggingMiddleware(kitlog.With(log.MiddlewareLogger, "method", s))(e)

	return e
}

// authMiddleware filter unauthorized requests, so next service is not called
func authMiddleware(authSvc service.AuthService) endpoint.Middleware {

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req := request.(TopContributorsRequest)

			i, err := authSvc.IsValidToken(ctx, req.Token)
			if err != nil {
				return
			}

			if !i {
				return nil, service.ErrUnauthorized
			}

			return next(ctx, request)
		}
	}
}
