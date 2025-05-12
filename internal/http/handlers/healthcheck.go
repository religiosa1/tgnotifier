package handlers

import (
	"net/http"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
)

type Healthcheck struct {
	Bot *tgnotifier.Bot
}

func (h Healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())
	if _, err := h.Bot.GetMe(logger); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
