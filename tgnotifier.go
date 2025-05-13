package tgnotifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Absolute maximum length of message body in bytes.
// Please notice, that TG applies 4096 **character** limit, 9500 is the limit
// for body length for formatting. So you still can get a Entities_too_long error
// on a message body shorter than this value, if the amount of characcters is more
// than 4096
// https://stackoverflow.com/questions/68768069/telegram-error-badrequest-entities-too-long-error-when-trying-to-send-long-ma
const MaxMsgLen int = 9500

type ParseMode = string

const (
	ParseModeMD       ParseMode = "MarkdownV2"
	ParseModeHTML     ParseMode = "HTML"
	ParseModeMDLegacy ParseMode = "Markdown"
)

func IsValidParseMode(parseMode string) bool {
	switch parseMode {
	case ParseModeMD, ParseModeHTML, ParseModeMDLegacy:
		return true
	default:
		return false
	}
}

var (
	ErrEmptyRecipients      = errors.New("empty recipients list")
	ErrInvalidMessageLength = errors.New("invalid message length")
	ErrInvalidParseMode     = errors.New("invalid parseMode value")
	ErrNotABot              = errors.New("we're not a bot according to getMe")
)

// Error responses returned from Telegram API
type TgApiError struct {
	TgCode      int
	Method      string
	Description string
}

func (e TgApiError) Error() string {
	return fmt.Sprintf("error during the TG API call to '%s' (%d): %s", e.Method, e.TgCode, e.Description)
}

const DefaultTimeout time.Duration = 30 * time.Second

type Bot struct {
	token      string
	httpClient *http.Client
}

func New(token string) *Bot {
	return NewWithClient(token, &http.Client{Timeout: DefaultTimeout})
}

func NewWithClient(token string, client *http.Client) *Bot {
	return &Bot{token, client}
}

//==============================================================================

func (bot *Bot) SendMessage(message string, parseMode ParseMode, recipients []string) error {
	return bot.SendMessageWithContext(context.Background(), message, parseMode, recipients)
}
func (bot *Bot) SendMessageWithContext(
	ctx context.Context,
	message string,
	parseMode ParseMode,
	recipients []string,
) error {
	if l := len(message); l < 1 || l >= MaxMsgLen {
		return ErrInvalidMessageLength
	}
	if parseMode != "" && !IsValidParseMode(parseMode) {
		return ErrInvalidParseMode
	}
	if len(recipients) == 0 {
		return ErrEmptyRecipients
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	errCh := make(chan error, len(recipients))
	var wg sync.WaitGroup
	wg.Add(len(recipients))

	for _, chatId := range recipients {
		go func(chatId string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}

			payload := SendMessagePayload{
				ChatId:    chatId,
				Text:      message,
				ParseMode: parseMode,
			}
			if err := bot.sendMessage(ctx, payload); err != nil {
				errCh <- err
			}
		}(chatId)
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
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

func (bot *Bot) sendMessage(ctx context.Context, payload SendMessagePayload) error {
	const method string = "sendMessage"
	endpointUrl := bot.methodUrl(method)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding the sendMessage body: %w", err)
	}
	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequestWithContext(ctx, "POST", endpointUrl, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bot.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	var sendMessageResponse BotResponse[struct{}]
	if err := json.NewDecoder(resp.Body).Decode(&sendMessageResponse); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if !sendMessageResponse.Ok {
		return TgApiError{sendMessageResponse.ErrorCode, method, sendMessageResponse.Description}
	}
	return nil
}

//==============================================================================

// @see https://core.telegram.org/bots/api#user
type GetMeResponse struct {
	Id        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	// optionals:
	LastName                string `json:"last_name,omitempty"`
	Username                string `json:"username,omitempty"`
	LanguageCode            string `json:"language_code,omitempty"`
	IsPremium               bool   `json:"is_premium,omitempty"`
	AddedToAttachmentMenu   bool   `json:"added_to_attachment_menu,omitempty"`
	CanJoinGroups           bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages,omitempty"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries,omitempty"`
}

func (bot *Bot) GetMe() (*GetMeResponse, error) {
	return bot.GetMeWithContext(context.Background())
}

func (bot *Bot) GetMeWithContext(ctx context.Context) (*GetMeResponse, error) {
	const method string = "getMe"
	endpointUrl := bot.methodUrl(method)
	req, err := http.NewRequestWithContext(ctx, "GET", endpointUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating bot request: %w", err)
	}

	resp, err := bot.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending bot request: %w", err)
	}
	defer resp.Body.Close()

	var getMeResp BotResponse[GetMeResponse]
	if err := json.NewDecoder(resp.Body).Decode(&getMeResp); err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	if !getMeResp.Ok {
		return nil, TgApiError{getMeResp.ErrorCode, method, getMeResp.Description}
	}
	if !getMeResp.Result.IsBot {
		return nil, ErrNotABot
	}
	return &getMeResp.Result, nil
}

//==============================================================================

func (bot *Bot) methodUrl(method string) string {
	escapedToken := url.PathEscape(bot.token)
	escapedMethod := url.PathEscape(method)
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", escapedToken, escapedMethod)
}
