package restchi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aawadallak/go-core-kit/plugin/rest"
	"github.com/go-chi/chi/v5"
)

// Server defines the interface for an HTTP server.
type Server interface {
	Start(ctx context.Context) error
	Handler() http.Handler
}

const (
	defaultAddr = ":8080" // Default address if not specified
)

// Option defines a function to configure the server.
type Option func(*server)

// WithAddr sets the server address (e.g., ":8080").
func WithAddr(addr string) Option {
	return func(s *server) {
		s.addr = addr
	}
}

// WithMiddlewares adds global middlewares to the server.
func WithMiddlewares(middlewares ...rest.Middleware) Option {
	return func(s *server) {
		s.middlewares = append(s.middlewares, middlewares...)
	}
}

// WithRoutes adds a list of routes to the server.
func WithRoutes(routes rest.Routes) Option {
	return func(s *server) {
		s.routes = append(s.routes, routes...)
	}
}

// WithRoute adds a single route to the server.
func WithRoute(route rest.Router) Option {
	return func(s *server) {
		s.routes = append(s.routes, route)
	}
}

// WithPrefix sets a global prefix for all routes.
func WithPrefix(prefix string) Option {
	return func(s *server) {
		s.prefix = prefix
	}
}

// WithRouteGroup adds a group of routes under a specific prefix with optional middlewares.
func WithRouteGroup(prefix string, middlewares []rest.Middleware, routes rest.Routes) Option {
	return func(s *server) {
		s.mux.Route(prefix, func(r chi.Router) {
			for _, mw := range middlewares {
				r.Use(adaptMiddleware(mw))
			}
			for _, route := range routes {
				r.Method(route.Method, route.Pattern, route.Handler)
			}
		})
	}
}

// WithLogger adds a simple request logging middleware.
func WithLogger(logger *log.Logger) Option {
	return func(s *server) {
		s.mux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				next.ServeHTTP(w, r)
				logger.Printf("%s %s - %s", r.Method, r.URL.Path, time.Since(start))
			})
		})
	}
}

// server is the internal implementation of the Server interface.
type server struct {
	mux         *chi.Mux
	middlewares []rest.Middleware
	routes      rest.Routes
	prefix      string
	addr        string
}

// adaptMiddleware converts a rest.Middleware to Chi's middleware signature.
func adaptMiddleware(mw rest.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return mw(next)
	}
}

// applyMiddlewares applies global middlewares to the router.
func (s *server) applyMiddlewares() {
	for _, mw := range s.middlewares {
		s.mux.Use(adaptMiddleware(mw))
	}
}

// registerRoutes registers all routes with the router, respecting the prefix.
func (s *server) registerRoutes() {
	s.mux.Route(s.prefix, func(r chi.Router) {
		for _, route := range s.routes {
			middlewares := make([]func(http.Handler) http.Handler, 0, len(route.Middlewares))
			for _, mw := range route.Middlewares {
				middlewares = append(middlewares, adaptMiddleware(mw))
			}

			r.With(middlewares...).Method(route.Method, route.Pattern, route.Handler)
		}
	})
}

// Start launches the server and respects the provided context for shutdown.
func (s *server) Start(ctx context.Context) error {
	if s.addr == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

// Handler returns the underlying HTTP handler for testing or composition.
func (s *server) Handler() http.Handler {
	return s.mux
}

// NewServer creates a new Chi-based HTTP server with the provided options.
func NewServer(opts ...Option) Server {
	srv := initializeServer()
	configureServer(srv, opts)
	return srv
}

// initializeServer sets up the base server structure.
func initializeServer() *server {
	return &server{
		mux: chi.NewRouter(),
	}
}

// configureServer applies options and initializes the server.
func configureServer(srv *server, opts []Option) {
	// Set default address, overridden by environment variable if present
	srv.addr = defaultAddr
	if port := os.Getenv("HTTP_PORT"); port != "" {
		srv.addr = ":" + port
	}

	// Apply all options
	for _, opt := range opts {
		opt(srv)
	}

	// Apply middlewares and routes
	srv.applyMiddlewares()
	srv.registerRoutes()
}
