package http

import (
	"errors"
	"fmt"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/service"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sync"
)

var ErrClosedConn = errors.New("use of closed network connection")

// Server defines http server
type Server struct {
	port     int
	listener net.Listener
	svc      service.Service
	authSvc  service.AuthService
	appName  string
	mutex    sync.Mutex
}

// New instantiates http server
func New(port int, svc service.Service, auth service.AuthService, appName string) *Server {
	return &Server{
		port:    port,
		svc:     svc,
		authSvc: auth,
		appName: appName,
	}
}

// Run http server
func (s *Server) Run() error {
	log.Info("Starting HTTP Server on port ", s.port)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Errorf("Unexpected Error Listening, err %s ", err.Error())

		return err
	}

	s.setListener(ln)

	r := mux.NewRouter()
	r.Methods("GET").Path("/top-contributors/v1").Handler(
		s.makeTopContributorsHandler(s.svc, s.appName))
	r.Methods("GET").Path("/top-contributors/v2").Handler(
		s.makeTopContributorsHandler(s.svc, fmt.Sprintf("%s_v2",s.appName)))

	r.Methods("GET").Path("/auth/top-contributors/v1").Handler(
		s.makeAuthTopContributorsHandler(s.svc, s.authSvc, s.appName))

	r.Methods("POST").Path("/auth").Handler(s.makeAuthHandler(s.authSvc, s.appName))
	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	http.Handle("/", r)

	err = http.Serve(ln, nil)
	if e, ok := err.(*net.OpError); ok {
		if errors.Is(e.Err, ErrClosedConn) {
			return nil
		}
	}

	return err
}

// Terminate stop server
func (s *Server) Terminate() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_ = s.listener.Close()
}

func (s *Server) setListener(ln net.Listener) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.listener = ln
}
