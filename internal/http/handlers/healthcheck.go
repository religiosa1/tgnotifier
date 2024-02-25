package handlers

import (
	"log/slog"
	"net/http"
	"simple-tg-notifier/internal/bot"
	"simple-tg-notifier/internal/http/middleware"
)

func Healthcheck(bot *bot.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger, ok := r.Context().Value(middleware.LoggingContextKey("logger")).(*slog.Logger)
		if !ok {
			logger = slog.Default()
		}
		if err := bot.GetMe(logger); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}
