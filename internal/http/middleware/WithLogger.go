package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
)

type LogginContextKey string

const loggingContextRequestId = LogginContextKey("logging_context.request_id")
const logginContextLogger = LogginContextKey("logging_context.logger")

type LoggingContext struct {
	Logger    *slog.Logger
	RequestId string
}

func WithLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := ulid.Make().String()
			newLogger := logger.With(slog.String("request_id", id))

			ctx := context.WithValue(r.Context(), loggingContextRequestId, id)
			ctx = context.WithValue(ctx, logginContextLogger, newLogger)

			newLogger.Info("Incoming request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			t1 := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			w.Header().Set("X-Request-Id", id)

			next.ServeHTTP(rw, r.WithContext(ctx))
			newLogger.Info(
				"request completed",
				slog.String("duration", time.Since(t1).String()),
				slog.Int("status_code", rw.status),
			)
		})
	}
}

// Wrapper around the response writer, to capture the response code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(logginContextLogger).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}
	return logger
}

func GetRequestId(ctx context.Context) string {
	return ctx.Value(loggingContextRequestId).(string)
}
