package connections

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
)

type slogGormLogger struct {
	level logger.LogLevel
}

// NewSlogGormLogger returns a GORM logger.Interface that routes SQL query
// logs through slog, so they're structured JSON like the rest of the app's
// logs instead of GORM's default colorized plain text.
func NewSlogGormLogger() logger.Interface {
	return &slogGormLogger{level: logger.Warn}
}

func (l *slogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

func (l *slogGormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Info {
		slog.Info(msg, "args", args)
	}
}

func (l *slogGormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Warn {
		slog.Warn(msg, "args", args)
	}
}

func (l *slogGormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Error {
		slog.Error(msg, "args", args)
	}
}

func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent {
		return
	}
	sql, rows := fc()
	attrs := []any{"sql", sql, "rows", rows, "elapsed_ms", time.Since(begin).Milliseconds()}

	if err != nil && l.level >= logger.Error {
		slog.Error("gorm query failed", append(attrs, "error", err)...)
		return
	}
	if l.level >= logger.Info {
		slog.Debug("gorm query", attrs...)
	}
}
