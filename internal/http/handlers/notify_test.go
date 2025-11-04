package handlers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/http/handlers"
	"github.com/stretchr/testify/require"
)

func makeRequest(body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	return req, rec
}

func trimRespBody(resp *httptest.ResponseRecorder) string {
	return strings.Trim(resp.Body.String(), " \n")
}

func TestNotify_Success(t *testing.T) {
	mock := mockBot{}
	handler := handlers.Notify{
		Bot:        &mock,
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown"}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	expectedBody := `{"success":true}`
	require.Equal(t, expectedBody, trimRespBody(resp))
	require.Equal(t, []string{"user1"}, mock.LastCallRecipients)
}

func TestNotify_RecipientsThroughPayload(t *testing.T) {
	mock := mockBot{}
	handler := handlers.Notify{
		Bot:        &mock,
		Recipients: []string{},
	}

	body := `{"message": "hello", "parse_mode": "Markdown", "recipients":["payload_user"]}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	expectedBody := `{"success":true}`
	require.Equal(t, expectedBody, trimRespBody(resp))
	require.Equal(t, []string{"payload_user"}, mock.LastCallRecipients)
}

func TestNotify_RecipientsThroughPayloadOverrideDefault(t *testing.T) {
	mock := mockBot{}
	handler := handlers.Notify{
		Bot:        &mock,
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown", "recipients":["user2"]}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	expectedBody := `{"success":true}`
	require.Equal(t, expectedBody, trimRespBody(resp))
	require.Equal(t, []string{"user2"}, mock.LastCallRecipients)
}

func TestNotify_NoRecipients(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"no recipients", `{"message": "hello", "parse_mode": "Markdown"}`},
		{"empty array", `{"message": "hello", "parse_mode": "Markdown", "recipients": []}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockBot{}
			handler := handlers.Notify{
				Bot:        &mock,
				Recipients: []string{},
			}

			req, resp := makeRequest(tt.payload)

			handler.ServeHTTP(resp, req)

			require.Equal(t, http.StatusBadRequest, resp.Code)
			expectedBody := `{"success":false,"error":"Recipients list not provided in the request, and default recipient is not set in the config"}`
			require.Equal(t, expectedBody, trimRespBody(resp))
			require.Empty(t, mock.LastCallRecipients, "Expected no call to bot")
		})
	}
}

func TestNotify_EmptyRecipientsListOverridesDefaultAndErrorsOut(t *testing.T) {
	mock := mockBot{}
	handler := handlers.Notify{
		Bot:        &mock,
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown", "recipients": []}`

	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := `{"success":false,"error":"Recipients list not provided in the request, and default recipient is not set in the config"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
	require.Empty(t, mock.LastCallRecipients, "Expected no call to bot")
}

func TestNotify_MissingBody(t *testing.T) {
	handler := handlers.Notify{
		Bot:        &mockBot{},
		Recipients: []string{"user1"},
	}

	req, resp := makeRequest("")

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := `{"success":false,"error":"no body was provided"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_BadJSON(t *testing.T) {
	handler := handlers.Notify{
		Bot:        &mockBot{},
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello"` // malformed JSON
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := `{"success":false,"error":"unexpected EOF"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_SendMessageFails_Internal(t *testing.T) {
	handler := handlers.Notify{
		Bot:        &mockBot{Err: errors.New("some internal error")},
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown"}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
	expectedBody := `{"success":false,"error":"some internal error"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_TgApiError_BadRequest(t *testing.T) {
	err := tgnotifier.TgApiError{
		TgCode:      100,
		Method:      "sendMessage",
		Description: "bad token",
	}

	handler := handlers.Notify{
		Bot:        &mockBot{Err: err},
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown"}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := fmt.Sprintf(`{"success":false,"error":"%s"}`, err.Error())
	require.Equal(t, expectedBody, trimRespBody(resp))
}
