package main

import (
	"log/slog"
	"os"

	"github.com/P1coFly/CarInfoEM/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("CONFIG_PATH is not set")
		os.Exit(1)
	}
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting api-servies", "env", cfg.Env)
	log.Debug("cfg data", "data", cfg)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
