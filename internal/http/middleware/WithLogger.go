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
			newLogger := logger.With(slog.String("request_id", id))

			ctx := context.WithValue(r.Context(), LoggingContextKey("request_id"), id)
			ctx = context.WithValue(ctx, LoggingContextKey("logger"), newLogger)

			newLogger.Info("Incoming request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			t1 := time.Now()
			next(w, r.WithContext(ctx))
			defer func() {
				newLogger.Debug(
					"request completed",
					slog.String("duration", time.Since(t1).String()),
				)
			}()
		}
	}
}
