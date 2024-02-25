package middleware

import (
	"log/slog"
	"net/http"
	"simple-tg-notifier/internal/http/models"
)

func WithApiKeyAuth(key string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		resp := models.ResponsePayload{}

		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger, ok := ctx.Value(LoggingContextKey("logger")).(*slog.Logger)
			if !ok {
				logger = slog.Default()
			}

			if isValid, requestKey := authorizeKey(key, r); !isValid {
				w.Header().Set("Content-Type", "application/json")
				if requestKey == "" {
					w.WriteHeader(http.StatusUnauthorized)
					resp.Error = "Authentication Required"
					logger.Info("No authorization header is supplied")
					return
				}
				w.WriteHeader(http.StatusForbidden)
				resp.Error = "Authorization failed"
				logger.Info("Invalid authorization key supplied", slog.String("key", key))
				return
			}
			next(w, r)
		}
	}
}

func authorizeKey(key string, r *http.Request) (bool, string) {
	requestKey := r.Header.Get("x-api-key")
	if key == "" {
		cookieKey, _ := r.Cookie("X-API-KEY")
		requestKey = cookieKey.Value
	}
	return key == requestKey, requestKey
}
