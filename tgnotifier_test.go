package tgnotifier_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/religiosa1/tgnotifier"
)

//==============================================================================
// SendMessage

func TestSendMessageWithContext_Success(t *testing.T) {
	bot := newTestBot(t)

	url := getMockEndpoint("sendMessage")
	httpmock.RegisterResponder("POST", url,
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"ok": true,
		}),
	)

	err := bot.SendMessageWithContext(context.Background(), "hello", tgnotifier.ParseModeMD, []string{"123"})
	require.NoError(t, err)

	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 1, info["POST "+url])
}

func TestSendMessageWithContext_InvalidInputs(t *testing.T) {
	bot := newTestBot(t)

	tests := []struct {
		name       string
		message    string
		parseMode  tgnotifier.ParseMode
		recipients []string
		expected   error
	}{
		{"Empty message", "", "", []string{"123"}, tgnotifier.ErrMessageEmpty},
		{"Too long message", string(make([]byte, tgnotifier.MaxMsgLen+1)), "", []string{"123"}, tgnotifier.ErrMessageTooLong},
		{"Nullish recipients", "Hello", "", nil, tgnotifier.ErrRecipientsEmpty},
		{"Empty recipients", "Hello", "", []string{}, tgnotifier.ErrRecipientsEmpty},
		{"Invalid parseMode", "Hello", "BadMode", []string{"123"}, tgnotifier.ErrParseModeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bot.SendMessageWithContext(context.Background(), tt.message, tt.parseMode, tt.recipients)
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}

func TestSendMessageWithContext_HandlesHTTPErrors(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("POST", getMockEndpoint("sendMessage"),
		httpmock.NewErrorResponder(errors.New("network error")))

	err := bot.SendMessageWithContext(context.Background(), "Hello", "", []string{"123"})
	assert.ErrorContains(t, err, "network error")
}

func TestSendMessageWithContext_HandlesNon200(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("POST", getMockEndpoint("sendMessage"),
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"ok":          false,
			"error_code":  403,
			"description": "Forbidden",
		}))

	err := bot.SendMessageWithContext(context.Background(), "Hello", "", []string{"123"})
	assert.ErrorContains(t, err, "Forbidden")
	var tgErr tgnotifier.TgApiError
	assert.ErrorAs(t, err, &tgErr)
	assert.Equal(t, 403, tgErr.TgCode)
}

func TestSendMessageWithContext_ContextCanceled(t *testing.T) {
	bot := newTestBot(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bot.SendMessageWithContext(ctx, "Hi", "", []string{"123"})
	assert.ErrorIs(t, err, context.Canceled)
}

//==============================================================================
// GetMe

func TestGetMeWithContext_Success(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("GET", getMockEndpoint("getMe"),
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"id":         123,
				"is_bot":     true,
				"first_name": "MyBot",
			},
		}))

	resp, err := bot.GetMeWithContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(123), resp.Id)
	assert.Equal(t, "MyBot", resp.FirstName)
	assert.True(t, resp.IsBot)
}

func TestGetMeWithContext_HttpError(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("GET", getMockEndpoint("getMe"),
		httpmock.NewErrorResponder(errors.New("connection failed")))

	_, err := bot.GetMeWithContext(context.Background())
	assert.ErrorContains(t, err, "connection failed")
}

func TestGetMeWithContext_InvalidJSON(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("GET", getMockEndpoint("getMe"),
		httpmock.NewStringResponder(200, "{invalid-json"))

	_, err := bot.GetMeWithContext(context.Background())
	assert.ErrorContains(t, err, "invalid character")
}

func TestGetMeWithContext_NotOKResponse(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("GET", getMockEndpoint("getMe"),
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"ok":          false,
			"error_code":  401,
			"description": "unauthorized",
		}))

	_, err := bot.GetMeWithContext(context.Background())
	var tgErr tgnotifier.TgApiError
	assert.ErrorAs(t, err, &tgErr)
	assert.Equal(t, "unauthorized", tgErr.Description)
}

func TestGetMeWithContext_NotABot(t *testing.T) {
	bot := newTestBot(t)

	httpmock.RegisterResponder("GET", getMockEndpoint("getMe"),
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"is_bot": false,
			},
		}))

	_, err := bot.GetMeWithContext(context.Background())
	assert.ErrorIs(t, err, tgnotifier.ErrNotABot)
}

//==============================================================================
// utils

func newTestBot(t *testing.T) *tgnotifier.Bot {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	t.Cleanup(httpmock.DeactivateAndReset)

	bot, err := tgnotifier.NewWithClient("fake-token", client)
	require.NoError(t, err)
	return bot
}

func getMockEndpoint(method string) string {
	return fmt.Sprintf("https://api.telegram.org/botfake-token/%s", method)
}
