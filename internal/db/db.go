package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"

	"github.com/julientant/weightlogr-cli/internal/migrations"
)

// Open opens the SQLite database at path and runs migrations.
func Open(ctx context.Context, logger *slog.Logger, path string) (*sql.DB, error) {
	dsn := path + "?_journal_mode=WAL&_foreign_keys=on"
	logger.DebugContext(ctx, "opening sqlite connection", "path", path, "dsn", dsn)

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	logger.DebugContext(ctx, "pinging database")
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := runMigrations(ctx, logger, conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	logger.InfoContext(ctx, "database ready", "path", path)
	return conn, nil
}

func runMigrations(ctx context.Context, logger *slog.Logger, conn *sql.DB) error {
	logger.DebugContext(ctx, "loading embedded migration source")

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	driver, err := sqlite3.WithInstance(conn, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}

	logger.DebugContext(ctx, "running migrations")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, _ := m.Version()
	logger.InfoContext(ctx, "migrations applied", "version", version, "dirty", dirty)
	return nil
}
