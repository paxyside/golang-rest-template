package postgres

import (
	"context"
	"emperror.dev/errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type DB struct {
	Conn *pgxpool.Pool
	mu   sync.Mutex
}

// Close gracefully shuts down the database connection.
func (d *DB) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Conn != nil {
		d.Conn.Close()
	}
}

// Init initializes the database connection and applies migrations.
func Init(DBURI string) (*DB, error) {
	// Apply migrations
	if err := applyMigrations(DBURI); err != nil {
		return nil, errors.Wrap(err, "applyMigrations")
	}

	// Initialize database connection
	var db DB

	config, err := pgxpool.ParseConfig(DBURI)
	if err != nil {
		return nil, errors.Wrap(err, "pgxpool.ParseConfig")
	}

	config.MaxConns = int32(runtime.NumCPU())

	db.Conn, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, errors.Wrap(err, "pgxpool.NewWithConfig")
	}

	// Test the connection
	if err = testConnection(db.Conn); err != nil {
		return nil, errors.Wrap(err, "testConnection")
	}

	// Start a connection worker for automatic reconnection
	go connectionWorker(&db, DBURI)

	return &db, nil
}

// testConnection pings the database to ensure the connection is alive.
func testConnection(db *pgxpool.Pool) error {
	if err := db.Ping(context.Background()); err != nil {
		return errors.Wrap(err, "db.Ping")
	}

	return nil
}

// connectionWorker periodically checks the database connection and reconnects if necessary.
func connectionWorker(db *DB, dbURI string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := testConnection(db.Conn); err != nil {
			slog.Error("failed to ping database", slog.Any("error", err))
			db.mu.Lock()

			newConn, err := pgxpool.New(context.Background(), dbURI)
			if err != nil {
				slog.Error("failed to create new database connection", slog.Any("error", err))
			}

			db.Conn.Close()
			db.Conn = newConn
			db.mu.Unlock()
			slog.Info("reconnected to database")
		}
	}
}

// applyMigrations runs all pending database migrations.
func applyMigrations(DBURI string) error {
	m, err := migrate.New("file://migrations/", DBURI)
	if err != nil {
		return errors.Wrap(err, "migrate.New")
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "m.Up")
	}

	if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
		return errors.Errorf("failed to close migrate instance: sourceErr: %v, dbErr: %v", sourceErr, dbErr)
	}

	return nil
}
