package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier/internal/cmd"
	"github.com/religiosa1/tgnotifier/internal/config"
)

var version = ""

type CLI struct {
	Version     bool            `short:"v" help:"Show version information and exit."`
	GenerateKey cmd.GenerateKey `cmd:"" help:"Generate a key for the app HTTP API"`
	Serve       cmd.Serve       `cmd:"" default:"withargs" help:"Run HTTP server"`
	Send        cmd.Send        `cmd:"" help:"Send a message in the CLI mode"`
}

func (cmd *CLI) AfterApply() error {
	if cmd.Version {
		showVersion()
		os.Exit(0)
	}
	return nil
}

var cli CLI

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Vars{"default_config_path": config.DefaultConfigPath},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func showVersion() {
	// using version from ldflags first if defined
	if version != "" {
		fmt.Printf("%s\n", version)
		return
	}

	// if not defined, using version from buildinfo
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("Build information not available")
		return
	}

	var buildVersion string
	for _, dep := range info.Deps {
		if dep.Path == "github.com/religiosa1/tgnotifier" {
			buildVersion = dep.Version
			break
		}
	}

	switch buildVersion {
	case "":
		buildVersion = "(devel) unknown"
	case "(devel)":
		var commit string
		var dirty bool
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value
			}
			if setting.Key == "vcs.modified" {
				dirty = setting.Value == "true"
			}
		}
		buildVersion += " " + commit
		if dirty {
			buildVersion += " dirty"
		}
	default:
	}
	fmt.Printf("%s\n", buildVersion)
}
