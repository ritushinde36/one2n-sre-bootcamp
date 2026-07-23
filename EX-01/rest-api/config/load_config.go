package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}
	slog.Info("config loaded")
}
