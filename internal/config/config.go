package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string   `yaml:"env" env:"ENV" env-default:"local"`
	LogLevel   string   `yaml:"log_level" env:"BOT_LOG_LEVEL" env-default:"info"`
	BotToken   string   `yaml:"bot_token" env:"BOT_TOKEN" env-required:"true"`
	Recepients []string `yaml:"recepients" env:"BOT_RECEPIENTS" env-required:"true"`
	Address    string   `yaml:"address" env:"BOT_ADDR" env-default:"localhost:6000"`
	ApiKey     string   `yaml:"api_key" env:"BOT_API_KEY" env-required:"true"`
}

func MustLoad(configPath string) *Config {
	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatal("Error loading configuration: ", err)
		}
	} else if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Error loading configuration %s: %s", configPath, err)
	}

	if len(cfg.Recepients) == 0 {
		log.Fatal("No recepients were provided in the config, operation is impossible")
	}

	if l := len(cfg.ApiKey); l < 60 {
		log.Fatalf("Provided API Key's length must be at least 60 characters long, got %d", l)
	}

	switch cfg.LogLevel {
	case "":
		cfg.LogLevel = "info"
	case "info", "debug", "warn", "error":
		// everything is ok, no action needed
	default:
		log.Fatalf("Incorrect LogLevel value '%s'. Possible values are 'debug', 'info', 'warn', and 'error", cfg.LogLevel)
	}

	return &cfg
}
