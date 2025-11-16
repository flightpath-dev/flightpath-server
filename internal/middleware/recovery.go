package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery creates a panic recovery middleware
func Recovery(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic and stack trace
					logger.Printf("PANIC: %v\n%s", err, debug.Stack())

					// Return 500 error
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
