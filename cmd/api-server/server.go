package api_server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Server is a struct that wraps the http server
type Server struct {
	server *http.Server
	logger *slog.Logger
}

// NewServer creates a new Server with the given handler and options
func NewServer(handler http.Handler, opts ...ServerOption) *Server {
	// create a default server
	s := &Server{
		server: &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}

	// apply the options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ServerOption is a function that modifies the server
type ServerOption func(*Server)

// WithAddr sets the address of the server
func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.server.Addr = addr
	}
}

// WithLogger sets the logger of the server
func WithLogger(logger *slog.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// Start starts the server and listens for signals
func (s *Server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	// start the server
	go func() {
		s.logger.Info("Starting server", "addr", s.server.Addr)
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", err)
		}
	}()

	// wait for a signal and then
	// shutdown the server gracefully
	select {
	case <-stop:
	}

	s.logger.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
