package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/config"
	"github.com/religiosa1/tgnotifier/internal/http/handlers"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
)

var defaultConfigPath = "config.yml"

type CommonCliArgs struct {
	Config string `env:"BOT_CONFIG_PATH" default:"can" help:"Configuration file name" type:"path"`
}

func (c *CommonCliArgs) BeforeApply(ctx *kong.Context) error {
	if c.Config == "" {
		c.Config = defaultConfigPath
	}
	return nil
}

type SendCmd struct {
	CommonCliArgs
	Message string `arg:"" optional:"" help:"Message to send"`
	// TODO short flags
	ParseMode  string   `enum:"MarkdownV2,HTML,Markdown" default:"MarkdownV2" help:"Message parse mode"`
	Recipients []string `help:"Message recipients (defaults to value from config or env)"`
	BotToken   string   `help:"your bot token as given by botfather (defaults to value from config or env)"`
}

func (c *SendCmd) BeforeApply(ctx *kong.Context) error {
	if c.Message == "" {
		r := io.LimitReader(os.Stdin, int64(tgnotifier.MaxMsgLen))
		input, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		c.Message = string(input)
	}
	return nil
}

var CLI struct {
	GenerateKey struct{} `cmd:"" help:"Generate a key for the app HTTP API"`
	Serve       struct {
		CommonCliArgs
	} `cmd:"" default:"1" help:"Run HTTP server"`
	Send SendCmd `cmd:"" help:"Send a message in the CLI mode"`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "generate-key":
		generateApiKey()
		return
	case "send":
		send()
	case "serve":
		runServer(CLI.Serve.Config)
	default:
		panic(ctx.Command())
	}
}

func send() {
	// TODO change must load to simple load, because this mode doesn't really require a config
	cfg := config.MustLoad(CLI.Send.Config)
	if cfg != nil {
		if CLI.Send.BotToken == "" {
			CLI.Send.BotToken = cfg.BotToken
		}
		if len(CLI.Send.Recipients) == 0 {
			CLI.Send.Recipients = cfg.Recipients
		}
	}
	if CLI.Send.BotToken == "" {
		log.Fatalf("Bot token must be provided, through CLI flags, config or env variable")
	}
	if len(CLI.Send.Recipients) == 0 {
		log.Fatalf("Recipeints list must be provided, through CLI flags, config or env variable")
	}
	bot := tgnotifier.New(cfg.BotToken)

	if err := bot.SendMessage(CLI.Send.Message, CLI.Send.ParseMode, CLI.Send.Recipients); err != nil {
		log.Fatal(err)
	}
}

func generateApiKey() {
	key := make([]byte, 30)
	if _, err := rand.Read(key); err != nil {
		log.Fatal("Error while generating a random key", err)
	}
	fmt.Println(strings.ToUpper(hex.EncodeToString(key)))
}

func runServer(configPath string) {
	cfg := config.MustLoad(configPath)

	log := setupLogger(cfg.Env, cfg.LogLevel)
	bot := tgnotifier.New(cfg.BotToken)

	botInfo, err := bot.GetMe()
	if err != nil {
		log.Error("Error accessing the telegram API with the provided bot token", slog.Any("error", err))
		os.Exit(2)
	}
	log.Debug("Bot initialized", slog.Any("GetMeInfo", botInfo))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mux := http.NewServeMux()
		middlewares := middleware.Chain(
			middleware.WithLogger(log),
			middleware.WithApiKeyAuth(cfg.ApiKey),
		)
		mux.Handle("GET /", middlewares(handlers.Healthcheck{Bot: bot}))
		mux.Handle("POST /", middlewares(handlers.Notify{Bot: bot, Recipients: cfg.Recipients}))

		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Error("Error starting the server", slog.Any("error", err))
			os.Exit(1)
		}
	}()
	log.Info("Running bot http server", slog.String("address", cfg.Address), slog.Any("recipients", cfg.Recipients))

	<-done
	log.Info("Server closed")
}

func setupLogger(env string, logLevel string) *slog.Logger {
	const (
		envLocal = "local"
		envDev   = "development"
		envProd  = "production"
	)

	var logger *slog.Logger
	var programLevel = new(slog.LevelVar)
	programLevel.Set(strLogLevelToEnumValue(logLevel))
	hdlrOpts := &slog.HandlerOptions{Level: programLevel}
	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, hdlrOpts))
	case envDev:
		logger = slog.New((slog.NewJSONHandler(os.Stdout, hdlrOpts)))
	case envProd:
		logger = slog.New((slog.NewJSONHandler(os.Stdout, hdlrOpts)))
	default:
		log.Fatalf("Unknown environment type %s, available options are %s, %s %s", env, envLocal, envDev, envProd)
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
