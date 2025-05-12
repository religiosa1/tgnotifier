package middleware

import (
	"log/slog"
	"net/http"

	"github.com/religiosa1/tgnotifier/internal/http/models"
)

func WithApiKeyAuth(key string) Middleware {
	return func(next http.Handler) http.Handler {
		resp := models.ResponsePayload{}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := GetLogger(r.Context())

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
			next.ServeHTTP(w, r)
		})
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
