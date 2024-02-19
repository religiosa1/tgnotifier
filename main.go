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
	"simple-tg-notifier/internal/bot"
	"simple-tg-notifier/internal/config"
	"simple-tg-notifier/internal/http/handlers"
	"strings"
	"syscall"
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
	fmt.Println(cfg)

	log := setupLogger(cfg.Env)
	bot := bot.New(cfg.BotToken, cfg.Recepients)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /notify", handlers.Notify(log, bot, cfg.ApiKey))

		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Error("Error starting the server", err)
			os.Exit(1)
		}
	}()
	log.Info("Running bot http server", slog.String("address", cfg.Address))

	<-done
	log.Info("Server closed")
}

func setupLogger(env string) *slog.Logger {
	const (
		envLocal = "local"
		envDev   = "development"
		envProd  = "production"
	)

	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	case envDev:
		logger = slog.New((slog.NewJSONHandler(os.Stdout, nil)))
	case envProd:
		logger = slog.New((slog.NewJSONHandler(os.Stdout, nil)))
	default:
		log.Fatalf("Unknown environment type %s, available options are %s, %s %s", env, envLocal, envDev, envProd)
	}
	return logger
}
