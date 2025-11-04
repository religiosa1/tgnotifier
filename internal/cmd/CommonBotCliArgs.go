package cmd

import (
	"errors"

	"github.com/religiosa1/tgnotifier/internal/config"
)

type CommandInterface interface {
	MergeConfig(cfg config.Config)
	ValidatePostMerge() error
}

type CommonBotCliArgs struct {
	Config     string   `short:"c" help:"Configuration file path ($BOT_CONFIG_PATH)"`
	Recipients []string `short:"r" help:"Message recipients, comma separated (defaults to value from config or $BOT_RECIPIENTS)"`
	BotToken   string   `yaml:"bot_token" help:"Your bot token as given by botfather (defaults to value from config or $BOT_TOKEN)"`
}

func (cmd *CommonBotCliArgs) MergeConfig(cfg config.Config) {
	if len(cmd.Recipients) == 0 {
		cmd.Recipients = cfg.Recipients
	}
	if cmd.BotToken == "" {
		cmd.BotToken = cfg.BotToken
	}
}

func (cmd *CommonBotCliArgs) ValidatePostMerge() error {
	// this all is also validated in the Bot itself, it's here only for the more informative error message
	if cmd.BotToken == "" {
		return errors.New("bot_token must be provided through the CLI, config or environment variable")
	}
	return nil
}
