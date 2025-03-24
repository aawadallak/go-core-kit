package restchi

import "github.com/aawadallak/go-core-kit/plugin/rest"

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
