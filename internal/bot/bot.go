package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
)

type Bot struct {
	token      string
	recepients []string
	httpClient *http.Client
}

func New(token string, recepients []string) *Bot {
	return &Bot{token, recepients, &http.Client{}}
}

// @see https://core.telegram.org/bots/api#sendmessage
type SendMessagePayload struct {
	// Unique identifier for the target chat or username of the target channel
	ChatId string `json:"chat_id"`
	// Text of the message to be sent, 1-4096 characters after entities parsing
	Text string `json:"text"`
	// Mode for parsing entities in the message text
	ParseMode string `json:"parse_mode,omitempty"`
}

type SendMessageResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

const (
	ParseModeMD       = "MarkdownV2"
	ParseModeHTML     = "HTML"
	ParseModeMDLegacy = "Markdown"
)

func (bot *Bot) SendMessage(logger *slog.Logger, message string, parseMode string) error {
	if l := len(message); l < 1 || l > 4096 {
		return errors.New("invalid message length")
	}
	if err := validateParseMode(parseMode); err != nil {
		return err
	}
	errCh := make(chan error, len(bot.recepients))
	var wg sync.WaitGroup
	wg.Add(len(bot.recepients))
	for i := 0; i < len(bot.recepients); i++ {
		chatId := bot.recepients[0]
		logger := logger.With(slog.String("chat_id", chatId))
		go func() {
			err := bot.sendMessage(logger, message, chatId)
			if err != nil {
				errCh <- err
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if l := len(errs); l > 1 {
		return fmt.Errorf("errors occurred while sending messages: %v", errs)
	} else if l == 1 {
		return errs[0]
	}
	return nil
}

func (bot *Bot) sendMessage(logger *slog.Logger, message string, chatId string) error {
	logger.Debug("Sending the notification")
	endpointUrl := bot.methodUrl("sendMessage")
	body, err := json.Marshal(SendMessagePayload{
		ChatId:    chatId,
		Text:      message,
		ParseMode: ParseModeMD,
	})
	if err != nil {
		logger.Error("Error encoding the sendMessage body", err)
		return err
	}
	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", endpointUrl, bodyReader)
	if err != nil {
		logger.Error("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bot.httpClient.Do(req)
	if err != nil {
		logger.Error("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	var response SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error("Error reading response body:", err)
		return err
	}

	if !response.Ok {
		logger.Error("Failed to call the API method", slog.Int("StatusCode", resp.StatusCode), slog.String("description", response.Description))
		return errors.New(response.Description)
	}
	logger.Debug("Sent the notification", slog.Int("StatusCode", resp.StatusCode))
	return nil
}

func (bot *Bot) methodUrl(method string) string {
	escapedToken := url.PathEscape(bot.token)
	escapedMethod := url.PathEscape(method)
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", escapedToken, escapedMethod)
}

func validateParseMode(parseMode string) error {
	switch parseMode {
	case ParseModeMD, ParseModeHTML, ParseModeMDLegacy, "":
		return nil
	default:
		return errors.New("invalid parseMode value")
	}
}
