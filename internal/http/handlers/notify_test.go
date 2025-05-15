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
	handler := handlers.Notify{
		Bot:        &mockBot{},
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown"}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	expectedBody := `{"requestId":"","success":true}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_MissingBody(t *testing.T) {
	mockBot := &mockBot{}

	handler := handlers.Notify{
		Bot:        mockBot,
		Recipients: []string{"user1"},
	}

	req, resp := makeRequest("")

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := `{"requestId":"","success":false,"error":"no body was provided"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_BadJSON(t *testing.T) {
	mockBot := &mockBot{}

	handler := handlers.Notify{
		Bot:        mockBot,
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello"` // malformed JSON
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := `{"requestId":"","success":false,"error":"unexpected EOF"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_SendMessageFails_Internal(t *testing.T) {
	mockBot := &mockBot{
		Err: errors.New("some internal error"),
	}

	handler := handlers.Notify{
		Bot:        mockBot,
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown"}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
	expectedBody := `{"requestId":"","success":false,"error":"some internal error"}`
	require.Equal(t, expectedBody, trimRespBody(resp))
}

func TestNotify_TgApiError_BadRequest(t *testing.T) {
	err := tgnotifier.TgApiError{
		TgCode:      100,
		Method:      "sendMessage",
		Description: "bad token",
	}
	mockBot := &mockBot{
		Err: err,
	}

	handler := handlers.Notify{
		Bot:        mockBot,
		Recipients: []string{"user1"},
	}

	body := `{"message": "hello", "parse_mode": "Markdown"}`
	req, resp := makeRequest(body)

	handler.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	expectedBody := fmt.Sprintf(`{"requestId":"","success":false,"error":"%s"}`, err.Error())
	require.Equal(t, expectedBody, trimRespBody(resp))
}
