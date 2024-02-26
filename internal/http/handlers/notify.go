package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
	"github.com/religiosa1/tgnotifier/internal/http/models"
)

type RequestPayload struct {
	Message   string `json:"message"`
	ParseMode string `json:"parse_mode"`
}

func Notify(botInstance *tgnotifier.Bot) http.HandlerFunc {
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
				return
			}
			resp.Error = err.Error()
			logger.Info("Failed to decode the body", err)
			return
		}
		if payload.Message == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			resp.Error = "'message' field is required"
			logger.Info("No message field was provided")
			return
		}
		if err := botInstance.SendMessage(logger, payload.Message, payload.ParseMode); err != nil {
			if errors.Is(err, tgnotifier.ErrBot) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp.Error = err.Error()
			return
		}
		resp.Success = true
	}
}
