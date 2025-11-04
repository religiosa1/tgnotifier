package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

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
	var triedPaths []string
	pathExplicitlySet := configPath != ""

	if !pathExplicitlySet {
		configPath = os.Getenv(configPathEnvKey)
		if configPath != "" {
			triedPaths = append(triedPaths, fmt.Sprintf("%s (from %s)", configPath, configPathEnvKey))
		}
		if configPath == "" {
			// Try default config paths in order: user config, then global config
			defaultPaths := getDefaultConfigPaths()
			for _, path := range defaultPaths {
				triedPaths = append(triedPaths, path)
				if _, err := os.Stat(path); err == nil {
					configPath = path
					break
				}
			}
		}
	} else {
		triedPaths = append(triedPaths, fmt.Sprintf("%s (explicitly set)", configPath))
	}

	var cfg Config
	fileExists := false
	if configPath != "" {
		_, err := os.Stat(configPath)
		fileExists = !os.IsNotExist(err)
	}

	if !fileExists {
		if pathExplicitlySet {
			return cfg, fmt.Errorf("specified config file does not exist: %s", configPath)
		}
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return cfg, fmt.Errorf("error loading configuration from environment: %w\nTried config paths:\n  %s",
				err, formatTriedPaths(triedPaths))
		}
	} else if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return cfg, fmt.Errorf("error loading configuration file: %w", err)
	}
	return cfg, nil
}

func formatTriedPaths(paths []string) string {
	if len(paths) == 0 {
		return "none"
	}
	result := ""
	for i, path := range paths {
		if i > 0 {
			result += "\n  "
		}
		result += path
	}
	return result
}
