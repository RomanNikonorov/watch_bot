package lib

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5/middleware"
)

func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type ReadinessCheck func(context.Context) error

func Readyz(isReady *atomic.Value, checks ...ReadinessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		results := make(chan error, len(checks))
		for _, check := range checks {
			go func(check ReadinessCheck) {
				results <- check(r.Context())
			}(check)
		}
		for range checks {
			select {
			case err := <-results:
				if err == nil {
					continue
				}
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			case <-r.Context().Done():
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

// LoggerWithSkipPaths returns a logger middleware that skips logging for specified paths
func LoggerWithSkipPaths(skipPaths ...string) func(next http.Handler) http.Handler {
	skipPathsSet := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skipPathsSet[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for specified paths
			if _, skip := skipPathsSet[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			// Use chi's default logger for other paths
			middleware.Logger(next).ServeHTTP(w, r)
		})
	}
}
