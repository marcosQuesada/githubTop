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

func (s *Server) makeTopContributorsHandler(svc service.Service, a string) http.Handler {
	e := makeTopContributorsEndpoint(svc)

	return s.makeContributorsHandler(e, a, "top_contributors")
}

func (s *Server) makeAuthTopContributorsHandler(svc service.Service, authSvc service.AuthService, a string) http.Handler {
	e := makeTopContributorsEndpoint(svc)
	e = authMiddleware(authSvc)(e)

	return s.makeContributorsHandler(e, a, "auth_top_contributors")
}

func (s *Server) makeContributorsHandler(e endpoint.Endpoint, a, k string) http.Handler {
	opts := []httptransport.ServerOption{httptransport.ServerErrorEncoder(errorEncoder)}

	return httptransport.NewServer(
		buildMiddleware(a, k, e),
		topContributorsRequestDecoder,
		responseEncoder,
		opts...,
	)
}

func (s *Server) makeAuthHandler(svc service.AuthService, appName string) http.Handler {
	opts := []httptransport.ServerOption{httptransport.ServerErrorEncoder(errorEncoder)}

	return httptransport.NewServer(
		buildMiddleware(appName, "auth", makeAuthEndpoint(svc)),
		authDecoder,
		authResponseEncoder,
		opts...,
	)
}

func makeTopContributorsEndpoint(svc service.Service) endpoint.Endpoint {
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
