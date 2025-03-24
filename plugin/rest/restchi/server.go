package restchi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aawadallak/go-core-kit/plugin/rest"
	"github.com/go-chi/chi/v5"
)

const (
	defaultAddr = ":8080" // Default address if not specified
)

// Server defines the interface for an HTTP server.
type Server interface {
	Start(ctx context.Context) error
	Handler() http.Handler
}

// server is the internal implementation of the Server interface.
type server struct {
	mux         *chi.Mux
	middlewares []rest.Middleware
	routes      rest.Routes
	prefix      string
	addr        string
	// TODO: Add support for TLS
}

// applyGlobalMiddlewares applies global middlewares to the router.
func (s *server) applyGlobalMiddlewares() {
	for _, mw := range s.middlewares {
		s.mux.Use(mw)
	}
}

// registerRoutes registers all routes with the router, respecting the prefix.
func (s *server) registerRoutes() {
	if s.prefix == "" {
		s.prefix = "/"
	}

	s.mux.Route(s.prefix, func(r chi.Router) {
		for _, route := range s.routes {
			middlewares := make([]func(http.Handler) http.Handler, 0, len(route.Middlewares))
			for _, mw := range route.Middlewares {
				middlewares = append(middlewares, mw)
			}

			fullPath := s.prefix + route.Pattern
			log.Printf("[REST] - method=%s path=%s",
				route.Method,
				fullPath,
			)

			r.With(middlewares...).Method(route.Method, route.Pattern, route.Handler)
		}
	})
}

// Start launches the server and respects the provided context and OS signals for shutdown.
func (s *server) Start(ctx context.Context) error {
	if s.addr == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel for server errors
	errChan := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("[REST] - starting server on %s", s.addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for either a signal, context cancellation, or server error
	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		// Received OS signal
	case <-ctx.Done():
		// Context cancelled
	}

	// Perform graceful shutdown
	shutdownCtx, cancel := context.
		WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("[REST] - shutting down server...")
	return httpServer.Shutdown(shutdownCtx)
}

func (s *server) Handler() http.Handler {
	return s.mux
}

func NewServer(opts ...Option) Server {
	srv := initializeServer()
	configureServer(srv, opts)
	return srv
}

func initializeServer() *server {
	return &server{
		mux: chi.NewRouter(),
	}
}

func configureServer(srv *server, opts []Option) {
	srv.addr = defaultAddr
	if port := os.Getenv("HTTP_PORT"); port != "" {
		srv.addr = ":" + port
	}

	// Apply all options
	for _, opt := range opts {
		opt(srv)
	}

	if len(srv.routes) > 0 {
		srv.registerRoutes()
	}

	if len(srv.middlewares) > 0 {
		srv.applyGlobalMiddlewares()
	}
}
