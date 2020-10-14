package metrics

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	pro "github.com/prometheus/client_golang/prometheus"
	"time"
)

// InstrumentingMiddleware returns an endpoint middleware that add counters and histograms, tracks
// success as bool
func InstrumentingMiddleware(n, s string) endpoint.Middleware {
	success, fail, duration := GetMetricsEndpoint(n, s)
	duration = duration.With("method", s)

	return instrumentingMiddlewareWithMetrics(success, fail, duration)
}

func instrumentingMiddlewareWithMetrics(success, fail metrics.Counter, duration metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			defer func(begin time.Time) {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())

				if err != nil {
					fail.Add(1)
					return
				}
				success.Add(1)
			}(time.Now())

			return next(ctx, request)
		}
	}
}

// GetMetricsEndpoint metric middleware
func GetMetricsEndpoint(n, s string) (success, error metrics.Counter, duration metrics.Histogram) {
	success = prometheus.NewCounterFrom(pro.CounterOpts{
		Namespace: n,
		Subsystem: s,
		Name:      "request_success",
		Help:      "Total success request.",
	}, []string{})

	error = prometheus.NewCounterFrom(pro.CounterOpts{
		Namespace: n,
		Subsystem: s,
		Name:      "request_error",
		Help:      "Total error request.",
	}, []string{})

	duration = prometheus.NewSummaryFrom(pro.SummaryOpts{
		Namespace: n,
		Subsystem: s,
		Name:      "request_duration_seconds",
		Help:      "Request duration in seconds.",
	}, []string{"method", "success"})

	return
}
