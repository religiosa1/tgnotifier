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

type BotResponse[T any] struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
	Result      T      `json:"result,omitempty"`
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

	var sendMessageResponse BotResponse[struct{}]
	if err := json.NewDecoder(resp.Body).Decode(&sendMessageResponse); err != nil {
		logger.Error("Error reading response body:", err)
		return err
	}

	if !sendMessageResponse.Ok {
		logger.Error("Failed to call the API method", slog.Int("StatusCode", resp.StatusCode), slog.String("description", sendMessageResponse.Description))
		return errors.New(sendMessageResponse.Description)
	}
	logger.Debug("Sent the notification", slog.Int("StatusCode", resp.StatusCode))
	return nil
}

func validateParseMode(parseMode string) error {
	switch parseMode {
	case ParseModeMD, ParseModeHTML, ParseModeMDLegacy, "":
		return nil
	default:
		return errors.New("invalid parseMode value")
	}
}

//==============================================================================

// @see https://core.telegram.org/bots/api#user
type GetMeResponse struct {
	Id       int64  `json:"id"`
	IsBot    bool   `json:"is_bot"`
	FirsName string `json:"first_name"`
	// optionals:
	LastName                string `json:"last_name,omitempty"`
	Username                string `json:"username,omitempty"`
	LanguageCode            string `json:"language_code,omitempty"`
	IsPremium               bool   `json:"is_premium,omitempty"`
	AddedToAttachmentMenu   bool   `json:"added_to_attachment_menu,omitempty"`
	CanJoingGroups          bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages,omitempty"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries,omitempty"`
}

func (bot *Bot) GetMe(logger *slog.Logger) error {
	logger.Debug("getMe healthcheck request")
	endpointUrl := bot.methodUrl("getMe")
	req, err := http.NewRequest("GET", endpointUrl, nil)
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

	var getMeInfo BotResponse[GetMeResponse]
	if err := json.NewDecoder(resp.Body).Decode(&getMeInfo); err != nil {
		logger.Error("Error reading response body:", err)
		return err
	}
	if !getMeInfo.Result.IsBot {
		logger.Error("Unexpected response from getMe request, we're supposed to be a bot:", slog.Any("GetMeInfo", getMeInfo))
		return errors.New("unexpected response from getMe request")
	}
	logger.Info("Bot initialized", slog.Any("GetMeInfo", getMeInfo))
	return nil
}

//==============================================================================

func (bot *Bot) methodUrl(method string) string {
	escapedToken := url.PathEscape(bot.token)
	escapedMethod := url.PathEscape(method)
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", escapedToken, escapedMethod)
}
