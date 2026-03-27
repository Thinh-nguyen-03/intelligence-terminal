package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: migrate <up|down|status|reset>\n")
		os.Exit(1)
	}
	command := os.Args[1]

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	// Find migrations directory relative to the repo root
	migrationsDir := findMigrationsDir()

	goose.SetDialect("postgres")

	switch command {
	case "up":
		if err := goose.Up(db, migrationsDir); err != nil {
			slog.Error("migration up failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations applied successfully")
	case "down":
		if err := goose.Down(db, migrationsDir); err != nil {
			slog.Error("migration down failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migration rolled back")
	case "status":
		if err := goose.Status(db, migrationsDir); err != nil {
			slog.Error("migration status failed", "error", err)
			os.Exit(1)
		}
	case "reset":
		if err := goose.Reset(db, migrationsDir); err != nil {
			slog.Error("migration reset failed", "error", err)
			os.Exit(1)
		}
		slog.Info("all migrations rolled back")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: migrate <up|down|status|reset>\n", command)
		os.Exit(1)
	}
}

func findMigrationsDir() string {
	// Try paths relative to common working directories
	candidates := []string{
		"db/migrations",
		"../../db/migrations",
		filepath.Join("..", "..", "db", "migrations"),
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(dir)
			slog.Info("using migrations directory", "path", abs)
			return dir
		}
	}

	// Fall back to explicit env var
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}

	slog.Error("could not find migrations directory, set MIGRATIONS_DIR env var")
	os.Exit(1)
	return ""
}
