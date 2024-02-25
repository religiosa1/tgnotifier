package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LoggingContextKey string
type LoggingContext struct {
	Logger    *slog.Logger
	RequestId string
}

func WithLogger(logger *slog.Logger) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := uuid.NewString()

			ctx := context.WithValue(r.Context(), LoggingContextKey("logger"), logger)
			ctx = context.WithValue(ctx, LoggingContextKey("request_id"), id)

			logger = logger.With(slog.String("request_id", id))
			logger.Info("Incoming request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			t1 := time.Now()
			next(w, r.WithContext(ctx))
			defer func() {
				logger.Debug(
					"request completed",
					slog.String("duration", time.Since(t1).String()),
				)
			}()
		}
	}
}
