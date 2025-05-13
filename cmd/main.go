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
	Config string `short:"c" env:"BOT_CONFIG_PATH" help:"Configuration file name" type:"path"`
}

func (c *CommonCliArgs) BeforeApply(ctx *kong.Context) error {
	if c.Config == "" {
		c.Config = defaultConfigPath
	}
	return nil
}

type SendCmd struct {
	CommonCliArgs
	Message    string   `arg:"" optional:"" help:"Message to send. Read from STDIN if not specified"`
	ParseMode  string   `short:"m" enum:"MarkdownV2,HTML,Markdown" default:"MarkdownV2" help:"Message parse mode"`
	Recipients []string `short:"r" help:"Message recipients, comma separated (defaults to value from config or env)"`
	BotToken   string   `help:"Your bot token as given by botfather (defaults to value from config or env)"`
}

func (c *SendCmd) AfterApply(ctx *kong.Context) error {
	if c.Message == "" {
		// limiting to one extra bite from the allowed max, so we can error out from tgnotifier
		r := io.LimitReader(os.Stdin, int64(tgnotifier.MaxMsgLen+1))
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
	case "send", "send <message>":
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
	if len(CLI.Send.Recipients) == 0 {
		fmt.Fprintf(os.Stderr, "Recipeints list must be provided, through CLI flags, config or env variable")
		os.Exit(1)
	}
	bot, err := tgnotifier.New(cfg.BotToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing the bot: %s", err)
		os.Exit(2)
	}
	if err := bot.SendMessage(CLI.Send.Message, CLI.Send.ParseMode, CLI.Send.Recipients); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending the message: %s", err)
		os.Exit(3)
	}
}

func generateApiKey() {
	key := make([]byte, 30)
	if _, err := rand.Read(key); err != nil {
		fmt.Fprintf(os.Stderr, "Error while generating a random key: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(strings.ToUpper(hex.EncodeToString(key)))
}

func runServer(configPath string) {
	cfg := config.MustLoad(configPath)

	logger := setupLogger(cfg.Env, cfg.LogLevel)
	bot, err := tgnotifier.New(cfg.BotToken)
	if err != nil {
		logger.Error("Error creating a bot", slog.Any("error", err))
		os.Exit(1)
	}

	botInfo, err := bot.GetMe()
	if err != nil {
		logger.Error("Error accessing the telegram API with the provided bot token", slog.Any("error", err))
		os.Exit(2)
	}
	logger.Debug("Bot initialized", slog.Any("GetMeInfo", botInfo))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mux := http.NewServeMux()
		middlewares := middleware.Chain(
			middleware.WithLogger(logger),
			middleware.WithApiKeyAuth(cfg.ApiKey),
		)
		mux.Handle("GET /", middlewares(handlers.Healthcheck{Bot: bot}))
		mux.Handle("POST /", middlewares(handlers.Notify{Bot: bot, Recipients: cfg.Recipients}))

		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			logger.Error("Error starting the server", slog.Any("error", err))
			os.Exit(3)
		}
	}()
	logger.Info("Running bot http server", slog.String("address", cfg.Address), slog.Any("recipients", cfg.Recipients))

	<-done
	logger.Info("Server closed")
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
