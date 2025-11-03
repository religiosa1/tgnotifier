package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/religiosa1/tgnotifier/internal/http/models"
)

func WithApiKeyAuth(configKey string) Middleware {
	if configKey == "" {
		return noopHandler
	}
	ctComparer := newConstantTimeComparer(configKey)
	return func(next http.Handler) http.Handler {
		resp := models.ResponsePayload{}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := GetLogger(r.Context())
			requestKey := getRequestKey(r)

			if !ctComparer.Eq(requestKey) {
				w.Header().Set("Content-Type", "application/json")
				if requestKey == "" {
					resp.Error = "Authentication Required"
					logger.Info("No authorization header is supplied")
					w.WriteHeader(http.StatusUnauthorized)
				} else {
					resp.Error = "Authorization failed"
					logger.Info("Invalid authorization key supplied")
					w.WriteHeader(http.StatusForbidden)
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					logger.Error("Error while writing response to client", slog.Any("error", err))
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getRequestKey(r *http.Request) string {
	requestKey := r.Header.Get("x-api-key")
	if requestKey == "" {
		if cookieKey, err := r.Cookie("X-API-KEY"); err == nil {
			requestKey = cookieKey.Value
		}
	}
	return requestKey
}

// constantTimeComparer of string or []bytes values, with hashing of
// provided values, so we're comparing against the same values length
//
// @see https://github.com/golang/go/issues/18936
type constantTimeComparer struct {
	targetValueHash [32]byte
}

func newConstantTimeComparer(targetValue string) constantTimeComparer {
	return constantTimeComparer{sha256.Sum256([]byte(targetValue))}
}

func (c constantTimeComparer) Eq(value string) bool {
	valueHash := sha256.Sum256([]byte(value))
	return subtle.ConstantTimeCompare(c.targetValueHash[:], valueHash[:]) == 1

}

func noopHandler(next http.Handler) http.Handler {
	return next
}
