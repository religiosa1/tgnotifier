package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LogginContextKey string

const loggingContextRequestId = LogginContextKey("logging_context.request_id")
const logginContextLogger = LogginContextKey("logging_context.logger")

type LoggingContext struct {
	Logger    *slog.Logger
	RequestId string
}

func WithLogger(logger *slog.Logger) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := uuid.NewString()
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
			next(w, r.WithContext(ctx))
			newLogger.Debug(
				"request completed",
				slog.String("duration", time.Since(t1).String()),
			)
		}
	}
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
