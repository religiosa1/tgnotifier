package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier"
)

type Send struct {
	CommonBotCliArgs
	ParseMode string `short:"m" enum:"MarkdownV2,HTML,Markdown" default:"MarkdownV2" help:"Message parse mode"`
	Message   string `arg:"" optional:"" help:"Message to send. Read from STDIN if not specified"`
}

func (c *Send) AfterApply(ctx *kong.Context) error {
	if c.Message == "" {
		// limiting to one extra bite from the allowed max, so we can error out from tgnotifier
		r := io.LimitReader(os.Stdin, int64(tgnotifier.MaxMsgLen+1))
		input, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		c.Message = string(input)
	}
	return nil
}

func (cmd *Send) Run() error {
	if len(cmd.Recipients) == 0 {
		return fmt.Errorf("recipeints list must be provided, through CLI flags, config or env variable")
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
