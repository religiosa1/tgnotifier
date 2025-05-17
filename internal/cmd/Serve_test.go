package cmd_test

import (
	"testing"

	"github.com/religiosa1/tgnotifier/internal/cmd"
	"github.com/religiosa1/tgnotifier/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestServe_parseFlags(t *testing.T) {
	var cmd cmd.Serve
	p := newCliParserWithConfig(t, &cmd, test.MockConfig)
	_, err := p.Parse([]string{"-r", "1,2,3", "-c", "test-conf", "--bot-token", "test-token",
		"--log-type", "text",
		"--log-level", "warn",
		"--api-key", "qwerty",
		"127.5.3.1:3000",
	})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, []string{"1", "2", "3"}, cmd.Recipients)
	assert.Equal(t, "test-conf", cmd.Config)
	assert.Equal(t, "test-token", cmd.BotToken)
	assert.Equal(t, "text", cmd.LogType)
	assert.Equal(t, "warn", cmd.LogLevel)
	assert.Equal(t, "qwerty", cmd.ApiKey)
	assert.Equal(t, "127.5.3.1:3000", cmd.Address)
}

func TestServe_parseWithDefaultsFromConfig(t *testing.T) {
	var cmd cmd.Serve
	p := newCliParserWithConfig(t, &cmd, test.MockConfig)
	_, err := p.Parse([]string{})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, test.MockConfig.Recipients, cmd.Recipients)
	assert.Equal(t, test.MockConfig.BotToken, cmd.BotToken)
	assert.Equal(t, test.MockConfig.LogType, cmd.LogType)
	assert.Equal(t, test.MockConfig.LogLevel, cmd.LogLevel)
	assert.Equal(t, test.MockConfig.ApiKey, cmd.ApiKey)
	assert.Equal(t, test.MockConfig.Address, cmd.Address)
}

func TestServe_parseEnvOverridesDefaults(t *testing.T) {
	// config path uses a separate mechanics so isn't included here
	t.Setenv("BOT_RECIPIENTS", "7,8,9")
	t.Setenv("BOT_TOKEN", "test-token")
	t.Setenv("BOT_LOG_TYPE", "text")
	t.Setenv("BOT_LOG_LEVEL", "warn")
	t.Setenv("BOT_API_KEY", "test-api-key")
	t.Setenv("BOT_ADDR", "test-addr")

	var cmd cmd.Serve
	p := newCliParserWithConfig(t, &cmd, test.MockConfig)
	_, err := p.Parse([]string{})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, []string{"7", "8", "9"}, cmd.Recipients)
	assert.Equal(t, "test-token", cmd.BotToken)
	assert.Equal(t, "text", cmd.LogType)
	assert.Equal(t, "warn", cmd.LogLevel)
	assert.Equal(t, "test-api-key", cmd.ApiKey)
	assert.Equal(t, "test-addr", cmd.Address)
}

func TestServe_parsePriorityFlagEnvConfig(t *testing.T) {
	// config path uses a separate mechanics so isn't included here
	t.Setenv("BOT_TOKEN", "test-token")
	t.Setenv("BOT_LOG_TYPE", "text")

	var cmd cmd.Serve
	p := newCliParserWithConfig(t, &cmd, test.MockConfig)
	_, err := p.Parse([]string{
		"--bot-token", "super-token",
		"--recipients", "5,6,7",
	})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, []string{"5", "6", "7"}, cmd.Recipients) // flag overrides calue without env
	assert.Equal(t, "super-token", cmd.BotToken)             // flag overrides value with env
	assert.Equal(t, "text", cmd.LogType)                     // env overrides value without flag
	assert.Equal(t, test.MockConfig.LogLevel, cmd.LogLevel)  // config is still the defaul
	assert.Equal(t, test.MockConfig.ApiKey, cmd.ApiKey)
	assert.Equal(t, test.MockConfig.Address, cmd.Address)
}
