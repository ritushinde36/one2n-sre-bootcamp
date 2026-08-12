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

// goose_migrate_lock is a MySQL named lock (GET_LOCK/RELEASE_LOCK) shared by
// every cmd/migrate invocation against the same database, so concurrent
// callers (e.g. multiple docker-migrate/app instances starting at once)
// serialize instead of racing to apply the same pending migration - goose
// itself has no MySQL locking built in.
const migrationLockName = "goose_migrate_lock"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	goose.SetLogger(&slogGooseLogger{})

	if err := config.LoadConfig(); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

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

	ctx := context.Background()
	conn, err := acquireMigrationLock(ctx, db)
	if err != nil {
		slog.Error("failed to acquire migration lock", "error", err)
		os.Exit(1)
	}
	defer releaseMigrationLock(ctx, conn)

	if command == "up" {
		if err := adoptPreExistingSchema(db); err != nil {
			slog.Error("refusing to proceed: pre-existing students table does not match migration 00001", "error", err)
			os.Exit(1)
		}
	}

	if err := goose.RunContext(ctx, command, db, migrationsDir); err != nil {
		slog.Error("migration failed", "command", command, "error", err)
		os.Exit(1)
	}
}

// acquireMigrationLock blocks (up to 30s) until it holds a MySQL named lock,
// so only one cmd/migrate process at a time checks and applies pending
// migrations. GET_LOCK is scoped to the connection that acquired it, so it
// must be acquired and released on the same *sql.Conn - not *sql.DB, which
// may hand different queries to different underlying connections.
func acquireMigrationLock(ctx context.Context, db *sql.DB) (*sql.Conn, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("reserving a connection: %w", err)
	}
	var got int
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30)", migrationLockName).Scan(&got); err != nil {
		conn.Close()
		return nil, fmt.Errorf("running GET_LOCK: %w", err)
	}
	if got != 1 {
		conn.Close()
		return nil, fmt.Errorf("timed out waiting for another migration to finish")
	}
	return conn, nil
}

func releaseMigrationLock(ctx context.Context, conn *sql.Conn) {
	if _, err := conn.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", migrationLockName); err != nil {
		slog.Error("failed to release migration lock", "error", err)
	}
	conn.Close()
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
