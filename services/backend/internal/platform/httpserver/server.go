package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

type Server struct {
	server *http.Server
}

func New(cfg config.HTTPConfig, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:         cfg.Address(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start(errCh chan<- error) {
	go func() {
		err := s.server.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return
		}

		errCh <- err
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Address() string {
	return s.server.Addr
}
