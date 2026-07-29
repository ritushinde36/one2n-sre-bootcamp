package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"

	"github.com/ritushinde36/one2n-sre-bootcamp/config"
)

const migrationsDir = "migrations"

func main() {
	config.LoadConfig()

	dsn := os.Getenv("DSN")
	if dsn == "" {
		slog.Error("environment variable DSN is not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./cmd/migrate <up|down|status>")
		os.Exit(1)
	}
	command := os.Args[1]

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("failed to open database connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := goose.SetDialect("mysql"); err != nil {
		slog.Error("failed to set goose dialect", "error", err)
		os.Exit(1)
	}

	if err := goose.Run(command, db, migrationsDir); err != nil {
		slog.Error("migration failed", "command", command, "error", err)
		os.Exit(1)
	}
}
