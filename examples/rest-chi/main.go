package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aawadallak/go-core-kit/plugin/rest"
	"github.com/aawadallak/go-core-kit/plugin/rest/restchi"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	middlewares := []rest.Middleware{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("Middleware 1")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("Middleware 2")
				next.ServeHTTP(w, r)
			})
		},
	}

	routes := rest.Routes{
		{
			Method:      "GET",
			Pattern:     "/api/v1/hello",
			Middlewares: middlewares,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Hello, World!")
			},
		},
	}

	srv := restchi.NewServer(
		restchi.WithAddr(":8080"),
		restchi.WithRoutes(routes),
		// Optional prefix
		restchi.WithPrefix("/saas"),
	)

	if err := srv.Start(ctx); err != nil {
		panic(err)
	}
}
