package handlers

import (
	"net/http"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
)

func Healthcheck(bot *tgnotifier.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetLogger(r.Context())
		if _, err := bot.GetMe(logger); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}
