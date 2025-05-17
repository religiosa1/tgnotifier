package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier/internal/cmd"
	"github.com/religiosa1/tgnotifier/internal/config"
	"gopkg.in/yaml.v3"
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
		kong.Vars{"default_config_path": config.DefaultConfigPath},
	}
	parser, err := kong.New(cmd, options...)
	if err != nil {
		t.Fatalf("error creating a parser: %v", err)
	}
	configFileName := createConfigFile(t, cfg)
	return &parserWithConfig{parser, cmd, configFileName}
}

func createConfigFile(t *testing.T, cfg config.Config) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFilePath := filepath.Join(tmpDir, "config.yml")
	cfgFile, err := os.Create(tmpFilePath)
	if err != nil {
		t.Fatalf("failed to open tmp file: %v", err)
	}
	defer cfgFile.Close()
	err = yaml.NewEncoder(cfgFile).Encode(cfg)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	return tmpFilePath
}

var mockConfig = config.Config{
	Address:    "127.1.1.1:3333",
	BotToken:   "1234567890:dY8ityIPogXaUqVrgH62AANw1AwFMn4EbMC",
	ApiKey:     "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEAD",
	LogLevel:   "error",
	LogType:    "json",
	Recipients: []string{"227039625"},
}
