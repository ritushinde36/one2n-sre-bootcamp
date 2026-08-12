package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads environment variables from a .env file in the working
// directory, if one exists. A missing file is not an error - it just means
// the caller is relying on the process's environment variables directly
// (e.g. --env-file in Docker). The caller decides what to do with a
// non-nil error, including whether it's fatal.
func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		slog.Info("no .env file found, using environment variables as-is")
		return nil
	}
	slog.Info("config loaded")
	return nil
}

// RequireEnv reads key from the environment and errors if it's unset or
// empty, so callers don't each have to duplicate that check themselves.
func RequireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("environment variable %s is not set", key)
	}
	return val, nil
}
