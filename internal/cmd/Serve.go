package cmd

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/http/handlers"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
)

type Serve struct {
	CommonBotCliArgs
	LogType  string `env:"BOT_LOG_TYPE" yaml:"log_type" enum:"text,json" default:"text" help:"Logger output type"`
	LogLevel string `env:"BOT_LOG_LEVEL" yaml:"log_level" default:"info" help:"Minimum logging level"`
	Address  string `arg:"" optional:"" env:"BOT_ADDR" yaml:"address" default:"localhost:6000" help:"HTTP server listening address"`
	ApiKey   string `env:"BOT_API_KEY" yaml:"api_key" help:"API key, passed in 'x-api-key' header to authorize incoming requests"`
}

func (cmd *Serve) Run() error {
	logger := setupLogger(cmd.LogLevel, cmd.LogLevel)
	bot, err := tgnotifier.New(cmd.BotToken)
	if err != nil {
		logger.Error("Error creating a bot", slog.Any("error", err))
		return err
	}

	botInfo, err := bot.GetMe()
	if err != nil {
		logger.Error("Error accessing the telegram API with the provided bot token", slog.Any("error", err))
		return err
	}
	logger.Debug("Bot initialized", slog.Any("GetMeInfo", botInfo))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mux := http.NewServeMux()
		middlewares := middleware.Chain(
			middleware.WithLogger(logger),
			middleware.WithApiKeyAuth(cmd.ApiKey),
		)
		mux.Handle("GET /", middlewares(handlers.Healthcheck{Bot: bot}))
		mux.Handle("POST /", middlewares(handlers.Notify{Bot: bot, Recipients: cmd.Recipients}))

		if err := http.ListenAndServe(cmd.Address, mux); err != nil {
			logger.Error("Error starting the server", slog.Any("error", err))
			os.Exit(3)
		}
	}()
	logger.Info("Running bot http server", slog.String("address", cmd.Address), slog.Any("recipients", cmd.Recipients))

	<-done
	logger.Info("Server closed")
	return nil
}

func setupLogger(logType string, logLevel string) *slog.Logger {
	var logger *slog.Logger
	var programLevel = new(slog.LevelVar)
	programLevel.Set(strLogLevelToEnumValue(logLevel))
	hdlrOpts := &slog.HandlerOptions{Level: programLevel}
	switch logType {
	case "text":
		logger = slog.New(slog.NewTextHandler(os.Stdout, hdlrOpts))
	case "json":
		logger = slog.New((slog.NewJSONHandler(os.Stdout, hdlrOpts)))
	default:
		log.Fatalf("Unknown logger type %s", logLevel)
	}
	return logger
}

func strLogLevelToEnumValue(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		log.Fatalf("Unexpected log level %s", logLevel)
		return slog.LevelInfo
	}
}
