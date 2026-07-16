package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code cleanly
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Log structured request metrics using native log package
		log.Printf("[HTTP] %s %s | Status: %d | Duration: %s | IP: %s | UA: %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start).String(),
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
