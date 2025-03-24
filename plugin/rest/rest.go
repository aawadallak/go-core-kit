package rest

import "net/http"

type Middleware func(next http.Handler) http.Handler

type Routes []Router

type Router struct {
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	Middlewares []Middleware
}
