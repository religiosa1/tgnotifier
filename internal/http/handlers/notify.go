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

type Notify struct {
	Bot *tgnotifier.Bot
}

func (h Notify) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logger := middleware.GetLogger(r.Context())

	writeResponse := func(statusCode int, payload models.ResponsePayload) {
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.Error("Error encoding response", slog.Any("error", err))
		}
	}

	resp := models.ResponsePayload{RequestId: middleware.GetRequestId(r.Context())}

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			resp.Error = "{ message: string } body expected but none was supplied"
			logger.Info("No body was supplied")
			return
		}
		resp.Error = err.Error()
		logger.Info("Failed to decode the body", slog.Any("error", err))
		writeResponse(http.StatusBadRequest, resp)
		return
	}
	if payload.Message == "" {
		resp.Error = "'message' field is required"
		writeResponse(http.StatusUnprocessableEntity, resp)
		logger.Info("No message field was provided")
		return
	}
	if err := h.Bot.SendMessage(logger, payload.Message, payload.ParseMode); err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, tgnotifier.ErrBot) {
			code = http.StatusBadRequest
		}
		resp.Error = err.Error()
		writeResponse(code, resp)
		return
	}
	resp.Success = true
	writeResponse(200, resp)
}
