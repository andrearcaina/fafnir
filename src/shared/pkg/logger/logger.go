package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
)

type Logger struct {
	logger *slog.Logger
}

// New takes in a slog.HandlerOptions and returns a new Logger instance with a JSON handler that writes to standard output
func New(opts *slog.HandlerOptions) *Logger {
	// opts is passed to the JSON handler to configure its behavior, such as time format, level encoding, etc.
	// if opts is nil, the JSON handler will use default options
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo, // default log level is Info (change to debug if more verbose logging is needed)
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.logger.Log(ctx, level, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *Logger) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			l.Info(r.Context(), "HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
