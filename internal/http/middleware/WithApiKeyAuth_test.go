package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/religiosa1/tgnotifier/internal/http/middleware"
	"github.com/religiosa1/tgnotifier/internal/http/models"
)

const validKey = "supersecretapikey"

func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
}

func parseResponse(t *testing.T, rr *httptest.ResponseRecorder) models.ResponsePayload {
	t.Helper()
	var payload models.ResponsePayload
	err := json.Unmarshal(rr.Body.Bytes(), &payload)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	return payload
}

func TestApiKeyAuth_Success_Header(t *testing.T) {
	mw := middleware.WithApiKeyAuth(validKey)
	handler := mw(testHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("x-api-key", validKey)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rr.Code)
	}
}

func TestApiKeyAuth_Success_Cookie(t *testing.T) {
	mw := middleware.WithApiKeyAuth(validKey)
	handler := mw(testHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "X-API-KEY", Value: validKey})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rr.Code)
	}
}

func TestApiKeyAuth_MissingKey(t *testing.T) {
	mw := middleware.WithApiKeyAuth(validKey)
	handler := mw(testHandler())

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", rr.Code)
	}

	resp := parseResponse(t, rr)
	if !strings.Contains(resp.Error, "Authentication Required") {
		t.Errorf("Expected authentication error, got: %s", resp.Error)
	}
}

func TestApiKeyAuth_InvalidKey(t *testing.T) {
	mw := middleware.WithApiKeyAuth(validKey)
	handler := mw(testHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("x-api-key", "wrongkey")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", rr.Code)
	}

	resp := parseResponse(t, rr)
	if !strings.Contains(resp.Error, "Authorization failed") {
		t.Errorf("Expected authorization failure message, got: %s", resp.Error)
	}
}

func TestApiKeyAuth_Disabled(t *testing.T) {
	mw := middleware.WithApiKeyAuth("")
	handler := mw(testHandler())

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK with no API key required, got %d", rr.Code)
	}
}
