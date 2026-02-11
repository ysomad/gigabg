package httpserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultAddr         = ":80"
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 5 * time.Second
)

type Server struct {
	server *http.Server
	notify chan error
}

func New(ctx context.Context, h http.Handler, opts ...Option) *Server {
	srv := &http.Server{
		Addr:         defaultAddr,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		Handler:      h,
	}

	s := &Server{
		server: srv,
		notify: make(chan error, 1),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.start(ctx)
	return s
}

func (s *Server) start(ctx context.Context) {
	go func() {
		slog.InfoContext(ctx, "httpeserver: starting at "+s.server.Addr)
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "httpserver: shutting down at "+s.server.Addr)
	return s.server.Shutdown(ctx)
}

type Option func(*Server)

func WithPort(port int) Option {
	return func(s *Server) {
		s.server.Addr = net.JoinHostPort("", strconv.Itoa(port))
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}
