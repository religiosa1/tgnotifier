package bot

import (
	"fmt"
	"log"
	"net/url"
)

type Bot struct {
	token string
}

func New(token string) *Bot {
	return &Bot{token}
}

func (bot *Bot) SendMessage(message string) {
	endpointUrl := bot.methodUrl("sendMessage")
	log.Println("Mocking sendMessage call: ", endpointUrl, message)
}

func (bot *Bot) methodUrl(method string) string {
	escapedToken := url.PathEscape(bot.token)
	escapedMethod := url.PathEscape(method)
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", escapedToken, escapedMethod)
}
