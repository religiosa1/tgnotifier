package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier/internal/cmd"
)

type CLI struct {
	ShortVersion bool            `short:"v" help:"Show version and exit."`
	GenerateKey  cmd.GenerateKey `cmd:"" help:"Generate a key for the app HTTP API"`
	Serve        cmd.Serve       `cmd:"" default:"withargs" help:"Run HTTP server"`
	Send         cmd.Send        `cmd:"" help:"Send a message in the CLI mode"`
	Version      cmd.Version     `cmd:"" help:"Show version and additional config information"`
}

func (cmd *CLI) AfterApply() error {
	if cmd.ShortVersion {
		fmt.Println(cmd.Version.GetVersion())
		os.Exit(0)
	}
	return nil
}

var cli CLI

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
