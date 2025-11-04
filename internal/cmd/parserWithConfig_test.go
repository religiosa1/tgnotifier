package cmd_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier/internal/cmd"
	"github.com/religiosa1/tgnotifier/internal/config"
	"github.com/religiosa1/tgnotifier/internal/test"
)

type parserWithConfig struct {
	*kong.Kong
	cmd            cmd.CommandInterface
	configFileName string
}

func (p *parserWithConfig) Parse(args []string) (*kong.Context, error) {
	ctx, err := p.Kong.Parse(args)
	if err != nil {
		return ctx, err
	}
	loadedCfg, err := config.Load(p.configFileName)
	if err != nil {
		return ctx, fmt.Errorf("failed to load config: %w", err)
	}
	p.cmd.MergeConfig(loadedCfg)
	return ctx, err
}

func newCliParserWithConfig(t *testing.T, cmd cmd.CommandInterface, cfg config.Config) *parserWithConfig {
	t.Helper()
	options := []kong.Option{
		kong.Name("test"),
		kong.Exit(func(int) {
			t.Helper()
			t.Fatalf("unexpected exit()")
		}),
	}
	parser, err := kong.New(cmd, options...)
	if err != nil {
		t.Fatalf("error creating a parser: %v", err)
	}
	configFileName := test.CreateConfigFile(t, cfg)
	return &parserWithConfig{parser, cmd, configFileName}
}
