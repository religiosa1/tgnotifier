package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

var DefaultConfigPath = "config.yml"

const configPathEnvKey = "BOT_CONFIG_PATH"

type Config struct {
	// logger type: "text", "json"
	LogType string `yaml:"log_type" env:"BOT_LOG_TYPE" env-default:"text"`
	// logger minimum level: "debug", "info", "warn", "error"
	LogLevel string `yaml:"log_level" env:"BOT_LOG_LEVEL" env-default:"info"`
	// your bot token as given by botfather
	BotToken   string   `yaml:"bot_token" env:"BOT_TOKEN"`
	Recipients []string `yaml:"recipients" env:"BOT_RECIPIENTS"`
	Address    string   `yaml:"address" env:"BOT_ADDR" env-default:"localhost:6000"`
	// API key, passed in 'x-api-key' to authorize requests to the app
	ApiKey string `yaml:"api_key" env:"BOT_API_KEY"`
}

func Load(configPath string) (Config, error) {
	pathExplicitlySet := configPath != ""
	if !pathExplicitlySet {
		configPath = os.Getenv(configPathEnvKey)
		if configPath == "" {
			configPath = DefaultConfigPath
		}
	}
	var cfg Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if pathExplicitlySet {
			return cfg, fmt.Errorf("specified config file does not exist: %s", configPath)
		}
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return cfg, fmt.Errorf("error loading configuration from environment: %w", err)
		}
	} else if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return cfg, fmt.Errorf("error loading configuration file: %w", err)
	}
	return cfg, nil
}
