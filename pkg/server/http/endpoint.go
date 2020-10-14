package http

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"net/http"
)

func (s *Server) makeTopContributorsHandler(svc Service, namespace string) http.Handler {
	e := makeTopContributorsEndpoint(svc)

	return s.makeTopContributorsTransport(e, namespace, "top_contributors")
}

func (s *Server) makeAuthTopContributorsHandler(svc Service, authSvc service.AuthService, namespace string) http.Handler {
	e := makeTopContributorsEndpoint(svc)
	e = authMiddleware(authSvc)(e)

	return s.makeTopContributorsTransport(e, namespace, "auth_top_contributors")
}

func (s *Server) makeTopSearchedLocationsHandler(svc Service, namespace string) http.Handler {
	e := makeTopSearchedLocationsEndpoint(svc)

	return s.makeSearchedLocationsTransport(e, namespace, "top_searched_locations")
}

func (s *Server) makeTopContributorsTransport(e endpoint.Endpoint, namespace, metricKey string) http.Handler {
	opts := []httptransport.ServerOption{httptransport.ServerErrorEncoder(errorEncoder)}

	return httptransport.NewServer(
		buildMiddleware(namespace, metricKey, e),
		topContributorsRequestDecoder,
		responseEncoder,
		opts...,
	)
}

func (s *Server) makeSearchedLocationsTransport(e endpoint.Endpoint, namespace, metricKey string) http.Handler {
	opts := []httptransport.ServerOption{httptransport.ServerErrorEncoder(errorEncoder)}

	return httptransport.NewServer(
		buildMiddleware(namespace, metricKey, e),
		topSearchedLocationsRequestDecoder,
		responseEncoder,
		opts...,
	)
}

func (s *Server) makeAuthTransport(svc service.AuthService, namespace string) http.Handler {
	opts := []httptransport.ServerOption{httptransport.ServerErrorEncoder(errorEncoder)}

	return httptransport.NewServer(
		buildMiddleware(namespace, "auth", makeAuthEndpoint(svc)),
		authDecoder,
		authResponseEncoder,
		opts...,
	)
}

func makeTopContributorsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(TopContributorsRequest)
		if !ok {
			return nil, errors.New("unexpected request type")
		}
		r := provider.GithubTopRequest{
			City:    req.City,
			Size:    req.Size,
			Version: req.APIv,
			Sort:    req.Sort,
		}
		c, err := svc.GetTopContributors(ctx, r)
		if err != nil {
			log.Errorf("Unexpected error getting Top contributors, err %s", err)
		}

		return TopContributorsResponse{Top: c}, err
	}
}

func makeAuthEndpoint(s service.AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AuthRequest)
		t, err := s.Authorize(ctx, req.User, req.Pass)
		if err != nil {
			log.Errorf("Unexpected error on Authentication, err %s", err)
		}

		return AuthResponse{Token: t}, err
	}
}

func makeTopSearchedLocationsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(TopSearchedLocationsRequest)
		if !ok {
			return nil, errors.New("unexpected request type")
		}

		c, err := svc.GetTopSearchedLocations(ctx, req.Size)
		if err != nil {
			log.Errorf("Unexpected error getting Top contributors, err %s", err)
		}

		return TopSearchedLocationsResponse{Top: c}, err
	}
}
