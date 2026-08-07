package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("failed to load .env file", "error", err)
			os.Exit(1)
		}
		slog.Info("no .env file found, using environment variables as-is")
		return
	}
	slog.Info("config loaded")
}
