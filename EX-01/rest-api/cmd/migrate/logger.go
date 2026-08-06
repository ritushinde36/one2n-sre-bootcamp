package main

import (
	"fmt"
	"log/slog"
	"os"
)

// slogGooseLogger routes goose's own internal logging through slog, so it
// comes out as structured JSON like the rest of the app's logs instead of
// goose's default plain-text output via the standard log package.
type slogGooseLogger struct{}

func (l *slogGooseLogger) Printf(format string, v ...any) {
	slog.Info(fmt.Sprintf(format, v...))
}

func (l *slogGooseLogger) Fatalf(format string, v ...any) {
	slog.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}
