package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"

	"github.com/ritushinde36/one2n-sre-bootcamp/config"
)

const migrationsDir = "migrations"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	goose.SetLogger(&slogGooseLogger{})

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

	if command == "up" {
		if err := adoptPreExistingSchema(db); err != nil {
			slog.Error("refusing to proceed: pre-existing students table does not match migration 00001", "error", err)
			os.Exit(1)
		}
	}

	if err := goose.RunContext(context.Background(), command, db, migrationsDir); err != nil {
		slog.Error("migration failed", "command", command, "error", err)
		os.Exit(1)
	}
}

func adoptPreExistingSchema(db *sql.DB) error {
	version, err := goose.EnsureDBVersion(db)
	if err != nil {
		return fmt.Errorf("checking goose version: %w", err)
	}
	if version >= 1 {
		return nil // migration 00001 already tracked, nothing to adopt
	}

	var tableName string
	err = db.QueryRow("SHOW TABLES LIKE 'students'").Scan(&tableName)
	if err == sql.ErrNoRows {
		return nil // table doesn't exist yet, let goose create it normally
	}
	if err != nil {
		return fmt.Errorf("checking for existing students table: %w", err)
	}

	expected, err := expectedCreateTableSQL()
	if err != nil {
		return err
	}

	var actualTable, actualDDL string
	if err := db.QueryRow("SHOW CREATE TABLE students").Scan(&actualTable, &actualDDL); err != nil {
		return fmt.Errorf("reading existing students table schema: %w", err)
	}

	if normalizeSQL(actualDDL) != normalizeSQL(expected) {
		return fmt.Errorf("existing schema does not match migrations/00001_create_students_table.sql - manual review required\nexisting:\n%s\nexpected:\n%s", actualDDL, expected)
	}

	slog.Info("existing students table matches migration 00001 - marking as applied instead of recreating")
	if _, err := db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, true)"); err != nil {
		return fmt.Errorf("marking migration 00001 as applied: %w", err)
	}
	return nil
}

func expectedCreateTableSQL() (string, error) {
	content, err := os.ReadFile(migrationsDir + "/00001_create_students_table.sql")
	if err != nil {
		return "", fmt.Errorf("reading migration file: %w", err)
	}
	up := strings.Split(string(content), "-- +goose Down")[0]
	up = strings.Replace(up, "-- +goose Up", "", 1)
	up = strings.TrimSuffix(strings.TrimSpace(up), ";")
	return up, nil
}

func normalizeSQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
