package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// GetEnv resolves key the way Docker/Compose secrets expect: if <key>_FILE
// points at a file (e.g. a mounted secret), its contents win; otherwise
// falls back to the plain <key> env var.
func GetEnv(key string) string {
	if path := os.Getenv(key + "_FILE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to read secret file", "path", path, "error", err)
			os.Exit(1)
		}
		return strings.TrimSpace(string(data))
	}
	return os.Getenv(key)
}

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
