package config_test

import (
	"testing"

	"github.com/religiosa1/tgnotifier/internal/config"
	"github.com/religiosa1/tgnotifier/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ConfigFromFile(t *testing.T) {

	cfgName := test.CreateConfigFile(t, test.MockConfig)
	cfg, err := config.Load(cfgName)
	require.NoError(t, err)

	assert.Equal(t, test.MockConfig.LogType, cfg.LogType)
	assert.Equal(t, test.MockConfig.LogLevel, cfg.LogLevel)
	assert.Equal(t, test.MockConfig.BotToken, cfg.BotToken)
	assert.Equal(t, test.MockConfig.Recipients, cfg.Recipients)
	assert.Equal(t, test.MockConfig.Address, cfg.Address)
	assert.Equal(t, test.MockConfig.ApiKey, cfg.ApiKey)
}

func TestLoad_ConfigFromEnv(t *testing.T) {
	t.Setenv("BOT_LOG_TYPE", "text")
	t.Setenv("BOT_LOG_LEVEL", "info")
	t.Setenv("BOT_TOKEN", "env-token")
	t.Setenv("BOT_RECIPIENTS", "111,222")
	t.Setenv("BOT_ADDR", "0.0.0.0:8080")
	t.Setenv("BOT_API_KEY", "env-secret")

	cfg, err := config.Load("") // No file, should fallback to env
	require.NoError(t, err)

	assert.Equal(t, "text", cfg.LogType)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "env-token", cfg.BotToken)
	assert.Equal(t, []string{"111", "222"}, cfg.Recipients)
	assert.Equal(t, "0.0.0.0:8080", cfg.Address)
	assert.Equal(t, "env-secret", cfg.ApiKey)
}

func TestLoad_MissingExplicitFile(t *testing.T) {
	cfg, err := config.Load("non-existent-config.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specified config file does not exist")
	assert.Empty(t, cfg)
}

func TestLoad_MissingImplicitFile(t *testing.T) {
	cfg, err := config.Load("")
	assert.NoError(t, err)

	// Default config values:
	assert.Equal(t, "text", cfg.LogType)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "", cfg.BotToken)
	assert.Empty(t, cfg.Recipients)
	assert.Equal(t, "localhost:6000", cfg.Address)
	assert.Equal(t, "", cfg.ApiKey)
}

func TestLoad_EnvOverridesConfig(t *testing.T) {
	t.Setenv("BOT_ADDR", "powerman:5000")

	cfgName := test.CreateConfigFile(t, test.MockConfig)
	cfg, err := config.Load(cfgName)
	require.NoError(t, err)

	assert.Equal(t, test.MockConfig.Recipients, cfg.Recipients) // keeps value where not overridden
	assert.Equal(t, "powerman:5000", cfg.Address)               // overrides from env
}
