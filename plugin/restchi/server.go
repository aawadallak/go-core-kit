package restchi

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/aawadallak/go-core-kit/plugin/rest"
	"github.com/go-chi/chi/v5"
)

const (
	defaultAddr = ":8080" // Default address if not specified
)

// Server defines the interface for an HTTP server.
type Server interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Handler() http.Handler
}

// server is the internal implementation of the Server interface.
type server struct {
	mux         *chi.Mux
	httpServer  *http.Server // Store the http.Server instance
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
	fn := func(r chi.Router) {
		for _, route := range s.routes {
			middlewares := make([]func(http.Handler) http.Handler, 0, len(route.Middlewares))
			for _, mw := range route.Middlewares {
				middlewares = append(middlewares, mw)
			}

			fullPath := s.prefix + route.Pattern

			logger.Global().InfoS(
				"REST Server",
				logger.WithValue("method", route.Method),
				logger.WithValue("path", fullPath),
			)

			r.With(middlewares...).Method(route.Method, route.Pattern, route.Handler)
		}
	}

	if s.prefix != "" {
		s.mux.Route(s.prefix, fn)
	} else {
		fn(s.mux)
	}
}

// Start launches the server and listens for the provided context to be canceled.
func (s *server) Start(ctx context.Context) error {
	if s.httpServer.Addr == "" { // Use httpServer.Addr now
		return errors.New("server address cannot be empty")
	}

	// Channel for server errors
	errChan := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for the context to be canceled or a server error
	select {
	case err := <-errChan:
		// Server failed to start or encountered a critical error
		return err
	case <-ctx.Done():
		// Context was canceled by the main application,
		// signaling us to stop. This is not an error.
		return nil // Return nil, as shutdown will be handled by the Shutdown method
	}
}

// Shutdown gracefully shuts down the HTTP server.
func (s *server) Shutdown(_ context.Context) error {
	// Create a shutdown context with a timeout.
	// This context is separate from the one passed to Start.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Adjust timeout as needed
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err // Return the error if shutdown fails
	}

	return nil
}

func (s *server) Handler() http.Handler {
	return s.mux
}

func NewServer(opts ...Option) Server {
	srv := initializeServer()
	configureServer(srv, opts)

	srv.httpServer = &http.Server{
		Addr:           srv.addr,
		Handler:        srv.mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

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

	if len(srv.middlewares) > 0 {
		srv.applyGlobalMiddlewares()
	}

	if len(srv.routes) > 0 {
		srv.registerRoutes()
	}
}
