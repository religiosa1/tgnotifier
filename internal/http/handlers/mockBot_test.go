package handlers_test

import (
	"context"

	"github.com/religiosa1/tgnotifier"
)

type mockBot struct {
	Err                error
	GetMeResponse      tgnotifier.GetMeResponse
	LastCallRecipients []string
}

func (b *mockBot) SendMessage(message string, parseMode tgnotifier.ParseMode, recipients []string) error {
	return b.SendMessageWithContext(context.Background(), message, parseMode, recipients)
}

func (b *mockBot) SendMessageWithContext(
	ctx context.Context,
	message string,
	parseMode tgnotifier.ParseMode,
	recipients []string,
) error {
	b.LastCallRecipients = recipients
	return b.Err
}

func (b *mockBot) GetMe() (tgnotifier.GetMeResponse, error) {
	return b.GetMeWithContext(context.Background())
}

func (b *mockBot) GetMeWithContext(ctx context.Context) (tgnotifier.GetMeResponse, error) {
	return b.GetMeResponse, b.Err
}
