package cmd

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/config"
	"github.com/religiosa1/tgnotifier/internal/http/handlers"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
)

// We can't use enums, default values, etc. unless we implement a custom resolver
// in kong to apply config. And we can't do that either, until this issue is resolved
// as we need to know if config was set explicitely: https://github.com/alecthomas/kong/issues/365

type Serve struct {
	CommonBotCliArgs `embed:""`
	LogType          string `placeholder:"text" help:"Logger output type ($BOT_LOG_TYPE)"`
	LogLevel         string `placeholder:"info" help:"Minimum logging level ($BOT_LOG_LEVEL)"`
	Address          string `arg:"" optional:"" env:"BOT_ADDR" default:"localhost:6000" help:"HTTP server listening address ($BOT_ADDR)"`
	ApiKey           string `help:"API key, passed in 'x-api-key' header to authorize incoming requests ($BOT_API_KEY)"`
}

func (cmd *Serve) MergeConfig(cfg config.Config) {
	cmd.CommonBotCliArgs.MergeConfig(cfg)
	MergeValueInto(&cmd.LogType, cfg.LogType)
	MergeValueInto(&cmd.LogLevel, cfg.LogLevel)
	MergeValueInto(&cmd.Address, cfg.Address)
	MergeValueInto(&cmd.ApiKey, cfg.ApiKey)
}
func MergeValueInto[T comparable](target *T, source T) {
	var zero T
	if *target == zero {
		*target = source
	}
}

func (cmd *Serve) ValidatePostMerge() error {
	if err := cmd.CommonBotCliArgs.ValidatePostMerge(); err != nil {
		return err
	}
	if cmd.LogType != "text" && cmd.LogType != "json" {
		return errors.New(`incorrect value for log type, only "text" and "json" are supported`)
	}
	if len(cmd.Recipients) == 0 {
		return errors.New("recipients list must be provided through the CLI, config or environment variable")
	}
	return nil
}

func (cmd *Serve) Run() error {
	cfg, err := config.Load(cmd.Config)
	if err != nil {
		return err
	}
	cmd.MergeConfig(cfg)
	err = cmd.ValidatePostMerge()
	if err != nil {
		return err
	}

	logger := setupLogger(cmd.LogType, cmd.LogLevel)
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
		log.Fatalf("Unknown logger type %s", logType)
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
