package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"simple-tg-notifier/internal/bot"
	"simple-tg-notifier/internal/http/middleware"
	"simple-tg-notifier/internal/http/models"
)

type RequestPayload struct {
	Message   string `json:"message"`
	ParseMode string `json:"parse_mode"`
}

func Notify(bot *bot.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx := r.Context()
		logger, ok := ctx.Value(middleware.LoggingContextKey("logger")).(*slog.Logger)
		if !ok {
			logger = slog.Default()
		}
		id := ctx.Value(middleware.LoggingContextKey("request_id")).(string)

		resp := models.ResponsePayload{RequestId: id}
		defer func() {
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				logger.Error("Error encoding response", err)
			}
		}()

		var payload RequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if errors.Is(err, io.EOF) {
				resp.Error = "{ message: string } body expected but none was supplied"
				logger.Info("No body was supplied")
			} else {
				resp.Error = err.Error()
				logger.Info("Failed to decode the body", err)
			}
			return
		}
		if payload.Message == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			resp.Error = "'message' field is required"
			logger.Info("No message field was provided")
			return
		}
		err := bot.SendMessage(logger, payload.Message, payload.ParseMode)
		if err == nil {
			resp.Success = true
		} else {
			resp.Error = err.Error()
		}
	}
}
