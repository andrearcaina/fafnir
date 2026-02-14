package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
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

// FieldExtractor is a function that extracts fields from a gRPC request and response to be included in the logs.
// It returns a map of key-value pairs. If nil, no extra fields are logged.
type FieldExtractor func(fullMethod string, req interface{}, resp interface{}) map[string]any

func (l *Logger) NewGRPCLoggingInterceptor(extractor FieldExtractor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		l.Debug(ctx, "gRPC Request", "method", info.FullMethod, "payload", req)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := status.Code(err)

		args := []any{
			"method", info.FullMethod,
			"duration", duration.String(),
			"status_code", statusCode.String(),
		}

		if extractor != nil {
			fields := extractor(info.FullMethod, req, resp)
			for k, v := range fields {
				args = append(args, k, v)
			}
		}

		if err != nil {
			args = append(args, "error", err.Error())
			l.Error(ctx, "gRPC Request Failed", args...)
		} else {
			l.Info(ctx, "gRPC Request Processed", args...)
		}

		return resp, err
	}
}
