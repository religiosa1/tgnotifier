package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/config"
)

type Send struct {
	CommonBotCliArgs `embed:""`
	ParseMode        string `short:"m" placeholder:"MarkdownV2" help:"Message parse mode"`
	Message          string `arg:"" optional:"" help:"Message to send. Read from STDIN if not specified"`
}

func (cmd *Send) AfterApply(ctx *kong.Context) error {
	if cmd.Message == "" {
		// limiting to one extra bite from the allowed max, so we can error out from tgnotifier
		r := io.LimitReader(os.Stdin, int64(tgnotifier.MaxMsgLen+1))
		input, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		cmd.Message = string(input)
	}
	return nil
}

func (cmd *Send) Run() error {
	cfg, err := config.Load(cmd.Config)
	if err != nil {
		return err
	}
	cmd.MergeConfig(cfg)
	// we're not validating the Send struct, only common args, allowing bot to error out
	if err := cmd.ValidatePostMerge(); err != nil {
		return err
	}
	bot, err := tgnotifier.New(cmd.BotToken)
	if err != nil {
		return fmt.Errorf("error initializing the bot: %w", err)
	}
	if err := bot.SendMessage(cmd.Message, cmd.ParseMode, cmd.Recipients); err != nil {
		return fmt.Errorf("error sending the message: %w", err)
	}
	return nil
}
