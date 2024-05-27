package middleware

import "net/http"

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain applies a chain of middlewares to a handler.
func Chain(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := range middlewares {
		handler = middlewares[len(middlewares)-1-i](handler)
	}

	return handler
}
