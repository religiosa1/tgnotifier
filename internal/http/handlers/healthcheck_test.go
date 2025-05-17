package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/http/handlers"
	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	cases := []struct {
		Name          string
		GetMeResponse tgnotifier.GetMeResponse
		Err           error
		Want          int
	}{
		{"Alive", tgnotifier.GetMeResponse{Username: "bot_username"}, nil, http.StatusOK},
		{"Dead", tgnotifier.GetMeResponse{}, errors.New("test"), http.StatusInternalServerError},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			bot := &mockBot{
				Err:           tt.Err,
				GetMeResponse: tt.GetMeResponse,
			}
			handler := handlers.Healthcheck{Bot: bot}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp := httptest.NewRecorder()

			handler.ServeHTTP(resp, req)

			require.Equal(t, resp.Code, tt.Want)
		})
	}
}
