package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"simple-tg-notifier/internal/bot"

	"github.com/google/uuid"
)

type RequestPayload struct {
	Message string `json:"message"`
}

type ResponsePayload struct {
	RequestId string `json:"requestId"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

func Notify(logger *slog.Logger, bot *bot.Bot, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()
		w.Header().Set("Content-Type", "application/json")

		logger = logger.With(slog.String("request_id", id))
		logRequest(logger, r)

		resp := ResponsePayload{RequestId: id}
		defer func() {
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				logger.Error("Error encoding response", err)
			}
		}()

		if isValid, requestKey := authorizeKey(key, r); !isValid {
			if requestKey == "" {
				w.WriteHeader(http.StatusUnauthorized)
				resp.Error = "Authentication Required"
			} else {
				w.WriteHeader(http.StatusForbidden)
				resp.Error = "Authorization failed"
				logger.Info("Invalid authorization key supplied", slog.String("key", key))
			}
			return
		}
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
		resp.Success = true
		bot.SendMessage(payload.Message)
	}
}

func logRequest(logger *slog.Logger, r *http.Request) {
	logger.Info("Incoming request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
	)
}

func authorizeKey(key string, r *http.Request) (bool, string) {
	requestKey := r.Header.Get("x-api-key")
	if key == "" {
		cookieKey, _ := r.Cookie("X-API-KEY")
		requestKey = cookieKey.Value
	}
	return key == requestKey, requestKey
}
