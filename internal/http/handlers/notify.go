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
	Message   string               `json:"message"`
	ParseMode tgnotifier.ParseMode `json:"parse_mode"`
	// recipients override (uses config values, if not provided)
	Recipients []string `json:"recipients"`
}

type Notify struct {
	Bot        tgnotifier.BotInterface
	Recipients []string
}

func (h Notify) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	writeResponse := func(statusCode int, payload models.ResponsePayload) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.Error("Error encoding response", slog.Any("error", err))
		}
	}

	resp := models.ResponsePayload{}

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			resp.Error = "no body was provided"
			logger.Info("No body was provided")
			writeResponse(http.StatusBadRequest, resp)
			return
		}
		resp.Error = err.Error()
		logger.Info("Failed to decode the body", slog.Any("error", err))
		writeResponse(http.StatusBadRequest, resp)
		return
	}

	var recipients []string
	if payload.Recipients != nil {
		recipients = payload.Recipients
	} else {
		recipients = h.Recipients
	}
	if len(recipients) == 0 {
		resp.Error = "Recipients list not provided in the request, and default recipient is not set in the config"
		writeResponse(400, resp)
		return
	}

	if err := h.Bot.SendMessageWithContext(r.Context(), payload.Message, payload.ParseMode, recipients); err != nil {
		logger.Error("Error sending the notification", slog.Any("error", err))
		resp.Error = err.Error()
		writeResponse(mapSendMessageErrorToHttpCode(err), resp)
		return
	}
	resp.Success = true
	writeResponse(http.StatusOK, resp)
}

func mapSendMessageErrorToHttpCode(err error) int {
	var apiError tgnotifier.TgApiError
	if errors.As(err, &apiError) {
		return http.StatusBadRequest
	}
	if errors.Is(err, tgnotifier.ErrMessageTooLong) {
		return http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, tgnotifier.ErrMessageEmpty) || errors.Is(err, tgnotifier.ErrParseModeInvalid) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}
