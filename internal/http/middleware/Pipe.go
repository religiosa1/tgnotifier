package middleware

import (
	"net/http"
)

func Pipe(handler http.HandlerFunc, middlewares ...(func(next http.HandlerFunc) http.HandlerFunc)) http.HandlerFunc {
	resultingHandler := handler
	for _, middleware := range middlewares {
		resultingHandler = middleware(resultingHandler)
	}
	return resultingHandler
}
