package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"

	applog "github.com/julientant/weightlogr-cli/internal/logger"
	"github.com/julientant/weightlogr-cli/internal/migrations"
)

const DSNPragmas = "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"

// Open opens the SQLite database at path and runs migrations.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	logger := applog.FromContext(ctx)
	dsn := path + DSNPragmas
	logger.DebugContext(ctx, "opening sqlite connection", "path", path, "dsn", dsn)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	logger.DebugContext(ctx, "pinging database")
	if err := conn.PingContext(ctx); err != nil {
		if cerr := conn.Close(); cerr != nil {
			logger.ErrorContext(ctx, "close after ping failure", "error", cerr)
		}
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := runMigrations(ctx, conn); err != nil {
		if cerr := conn.Close(); cerr != nil {
			logger.ErrorContext(ctx, "close after migration failure", "error", cerr)
		}
		return nil, fmt.Errorf("migrate: %w", err)
	}

	logger.InfoContext(ctx, "database ready", "path", path)
	return conn, nil
}

func runMigrations(ctx context.Context, conn *sql.DB) error {
	logger := applog.FromContext(ctx)
	logger.DebugContext(ctx, "loading embedded migration source")

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(conn, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}

	logger.DebugContext(ctx, "running migrations")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, verr := m.Version()
	if verr != nil {
		logger.WarnContext(ctx, "could not read migration version", "error", verr)
	} else {
		logger.InfoContext(ctx, "migrations applied", "version", version, "dirty", dirty)
	}
	return nil
}
