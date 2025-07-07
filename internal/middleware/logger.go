package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader overrides the WriteHeader (of http.ResponseWriter) method to capture the status code
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logger is a middleware that logs the HTTP request and response details (this is a custom middleware using zap.Logger for structured logging)
func Logger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			start := time.Now()
			next.ServeHTTP(rec, r)
			duration := time.Since(start).Truncate(time.Millisecond)
			logLine := fmt.Sprintf(
				`%q from %s - %d in %s`,
				fmt.Sprintf("%s %s %s", r.Method, r.URL.String(), r.Proto),
				r.RemoteAddr,
				rec.status,
				duration,
			)
			logger.Sugar().Info(logLine)
		})
	}
}
