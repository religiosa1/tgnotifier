package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/religiosa1/tgnotifier"
	"github.com/religiosa1/tgnotifier/internal/config"
	"github.com/religiosa1/tgnotifier/internal/http/handlers"
	"github.com/religiosa1/tgnotifier/internal/http/middleware"
)

func main() {
	var apiKeyFlag bool
	var configPath string
	flag.BoolVar(&apiKeyFlag, "generate-key", false, "Generate API KEY")
	flag.StringVar(&configPath, "config", "", "Configuration file name")
	flag.Parse()

	if apiKeyFlag {
		generateApiKey()
	} else {
		runServer(configPath)
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
	bot := tgnotifier.New(cfg.BotToken, cfg.Recepients)

	botInfo, err := bot.GetMe(log)
	if err != nil {
		log.Error("Error accessing the telegram API with the provided bot token", err)
		os.Exit(2)
	}
	log.Debug("Bot initialized", slog.Any("GetMeInfo", botInfo))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /", middleware.Pipe(
			handlers.Healthcheck(bot),
			middleware.WithApiKeyAuth(cfg.ApiKey),
			middleware.WithLogger(log),
		))
		mux.HandleFunc("POST /", middleware.Pipe(
			handlers.Notify(bot),
			middleware.WithApiKeyAuth(cfg.ApiKey),
			middleware.WithLogger(log),
		))

		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Error("Error starting the server", err)
			os.Exit(1)
		}
	}()
	log.Info("Running bot http server", slog.String("address", cfg.Address), slog.Any("recepients", cfg.Recepients))

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
