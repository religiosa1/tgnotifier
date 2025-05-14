package main

import (
	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier/internal/cmd"
	"github.com/religiosa1/tgnotifier/internal/config"
)

var CLI struct {
	GenerateKey cmd.GenerateKey `cmd:"" help:"Generate a key for the app HTTP API"`
	Serve       cmd.Serve       `cmd:"" default:"withargs" help:"Run HTTP server"`
	Send        cmd.Send        `cmd:"" help:"Send a message in the CLI mode"`
}

func main() {
	ctx := kong.Parse(
		&CLI,
		kong.Vars{"default_config_path": config.DefaultConfigPath},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
