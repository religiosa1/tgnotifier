package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/religiosa1/tgnotifier/internal/config"
	"gopkg.in/yaml.v3"
)

func CreateConfigFile(t *testing.T, cfg config.Config) string {
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

var MockConfig = config.Config{
	Address:    "127.1.1.1:3333",
	BotToken:   "1234567890:dY8ityIPogXaUqVrgH62AANw1AwFMn4EbMC",
	ApiKey:     "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEAD",
	LogLevel:   "error",
	LogType:    "json",
	Recipients: []string{"227039625"},
}
