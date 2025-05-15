package handlers

import (
	"net/http"

	"github.com/religiosa1/tgnotifier"
)

type Healthcheck struct {
	Bot tgnotifier.BotInterface
}

func (h Healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := h.Bot.GetMeWithContext(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
