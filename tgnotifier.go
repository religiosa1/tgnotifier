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

// MaxMsgLen is the maximum allowed message body length in bytes.
//
// Note: Telegram limits messages to 4096 characters, but this value accounts for
// formatting and encoded entities. You may still receive an "entities too long"
// error if the message exceeds 4096 characters even though it's under MaxMsgLen.
//
// See: https://stackoverflow.com/questions/68768069/telegram-error-badrequest-entities-too-long-error-when-trying-to-send-long-ma
const MaxMsgLen int = 9500

type ParseMode = string

const (
	ParseModeMD       ParseMode = "MarkdownV2"
	ParseModeHTML     ParseMode = "HTML"
	ParseModeMDLegacy ParseMode = "Markdown"
)

// IsValidParseMode reports whether the given parse mode is valid.
func IsValidParseMode(parseMode string) bool {
	switch parseMode {
	case ParseModeMD, ParseModeHTML, ParseModeMDLegacy:
		return true
	default:
		return false
	}
}

var (
	ErrTokenEmpty       = errors.New("empty TG bot token")
	ErrRecipientsEmpty  = errors.New("empty recipients list")
	ErrMessageEmpty     = errors.New("tg message is empty")
	ErrMessageTooLong   = errors.New("tg message length exceeds maximum")
	ErrParseModeInvalid = errors.New("invalid parseMode value")
	ErrNotABot          = errors.New("we're not a bot according to getMe")
)

// TgApiError represents an error returned by the Telegram Bot API.
type TgApiError struct {
	TgCode      int
	Method      string
	Description string
}

func (e TgApiError) Error() string {
	return fmt.Sprintf("error during the TG API call to '%s' (%d): %s", e.Method, e.TgCode, e.Description)
}

// DefaultTimeout is the timeout duration for the default Bot http client
// (the one created with [New], not [NewWithClient])
const DefaultTimeout time.Duration = 30 * time.Second

type BotInterface interface {
	SendMessage(message string, parseMode ParseMode, recipients []string) error
	SendMessageWithContext(ctx context.Context, message string, parseMode ParseMode, recipients []string) error
	GetMe() (GetMeResponse, error)
	GetMeWithContext(ctx context.Context) (GetMeResponse, error)
}

// Bot is a Telegram notification bot.
type Bot struct {
	token      string
	httpClient *http.Client
}

// New wraps [NewWithClient] using the default http.Client with Timeout: [DefaultTimeout]
func New(token string) (*Bot, error) {
	return NewWithClient(token, &http.Client{Timeout: DefaultTimeout})
}

// NewWithClient creates a new instance of Bot with the provided
// BOT API token and http client instance
func NewWithClient(token string, client *http.Client) (*Bot, error) {
	if token == "" {
		return nil, ErrTokenEmpty
	}
	return &Bot{token, client}, nil
}

//==============================================================================

// SendMessage wraps [SendMessageWithContext] using context.Background.
func (bot *Bot) SendMessage(message string, parseMode ParseMode, recipients []string) error {
	return bot.SendMessageWithContext(context.Background(), message, parseMode, recipients)
}

// SendMessage sends TG message in a given parseMode to one or more recipients
//
// See: https://core.telegram.org/bots/api#sendmessage
func (bot *Bot) SendMessageWithContext(
	ctx context.Context,
	message string,
	parseMode ParseMode,
	recipients []string,
) error {
	l := len(message)
	if l > MaxMsgLen {
		return ErrMessageTooLong
	}
	if l <= 0 {
		return ErrMessageEmpty
	}
	if parseMode != "" && !IsValidParseMode(parseMode) {
		return ErrParseModeInvalid
	}
	if len(recipients) == 0 {
		return ErrRecipientsEmpty
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

			payload := sendMessagePayload{
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

type botResponse[T any] struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
	Result      T      `json:"result,omitempty"`
}

// https://core.telegram.org/bots/api#sendmessage
type sendMessagePayload struct {
	ChatId    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func (bot *Bot) sendMessage(ctx context.Context, payload sendMessagePayload) error {
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

	var sendMessageResponse botResponse[struct{}]
	if err := json.NewDecoder(resp.Body).Decode(&sendMessageResponse); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if !sendMessageResponse.Ok {
		return TgApiError{sendMessageResponse.ErrorCode, method, sendMessageResponse.Description}
	}
	return nil
}

//==============================================================================

// GetMeResponse represents response payload of TG getMe endpoint -- the bot user.
//
// See: https://core.telegram.org/bots/api#user
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
	CanConnectToBusiness    bool   `json:"can_connect_to_business,omitempty"`
	HasMainWebApp           bool   `json:"has_main_web_app,omitempty"`
}

// GetMe wraps [GetMeWithContext] using context.Background.
func (bot *Bot) GetMe() (GetMeResponse, error) {
	return bot.GetMeWithContext(context.Background())
}

// GetMeWithContext calls `getMe` telegram endpoint for testing token and returns its response
//
// See: https://core.telegram.org/bots/api#getme
func (bot *Bot) GetMeWithContext(ctx context.Context) (GetMeResponse, error) {
	var getMeResp botResponse[GetMeResponse]

	const method string = "getMe"
	endpointUrl := bot.methodUrl(method)
	req, err := http.NewRequestWithContext(ctx, "GET", endpointUrl, nil)
	if err != nil {
		return getMeResp.Result, fmt.Errorf("error creating bot request: %w", err)
	}

	resp, err := bot.httpClient.Do(req)
	if err != nil {
		return getMeResp.Result, fmt.Errorf("error sending bot request: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&getMeResp); err != nil {
		return getMeResp.Result, fmt.Errorf("error reading response body: %w", err)
	}
	if !getMeResp.Ok {
		return getMeResp.Result, TgApiError{getMeResp.ErrorCode, method, getMeResp.Description}
	}
	if !getMeResp.Result.IsBot {
		return getMeResp.Result, ErrNotABot
	}
	return getMeResp.Result, nil
}

//==============================================================================

func (bot *Bot) methodUrl(method string) string {
	escapedToken := url.PathEscape(bot.token)
	escapedMethod := url.PathEscape(method)
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", escapedToken, escapedMethod)
}
