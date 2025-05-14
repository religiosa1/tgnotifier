package cmd

import (
	"errors"

	"github.com/religiosa1/tgnotifier/internal/config"
)

type CommonBotCliArgs struct {
	Config     string   `placeholder:"${default_config_path}" short:"c" help:"Configuration file name ($BOT_CONFIG_PATH)"`
	Recipients []string `short:"r" help:"Message recipients, comma separated (defaults to value from config or $BOT_RECIPIENTS)"`
	BotToken   string   `yaml:"bot_token" help:"Your bot token as given by botfather (defaults to value from config or $BOT_TOKEN)"`
}

func (cmd *CommonBotCliArgs) MergeConfig(cfg config.Config) {
	if len(cmd.Recipients) != 0 {
		cmd.Recipients = cfg.Recipients
	}
	if cmd.BotToken == "" {
		cmd.BotToken = cfg.BotToken
	}
}

func (cmd *CommonBotCliArgs) ValidatePostMerge() error {
	if cmd.BotToken == "" {
		return errors.New("bot_token must be provided through the CLI, config or environment variable")
	}
	return nil
}
