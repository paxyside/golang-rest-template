package postgres

import (
	"context"
	"emperror.dev/errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"sync"
	"time"
)

type DBExecutor interface {
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Close()
}

type DB struct {
	conn *pgxpool.Pool
	mu   sync.Mutex
}

func Init(ctx context.Context, dbURI string) (*DB, error) {
	if err := applyMigrations(dbURI); err != nil {
		return nil, errors.Wrap(err, "applyMigrations")
	}

	cfg, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return nil, errors.Wrap(err, "pgxpool.ParseConfig")
	}

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "pgxpool.NewWithConfig")
	}

	if err := testConnection(ctx, conn); err != nil {
		conn.Close()
		return nil, errors.Wrap(err, "testConnection")
	}

	db := &DB{conn: conn}

	go db.connectionWorker(ctx, dbURI)

	return db, nil
}

func (d *DB) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}
}

func (d *DB) connectionWorker(ctx context.Context, dbURI string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := testConnection(ctx, d.conn); err != nil {
				slog.Error("db ping failed", slog.Any("error", err))
				d.reconnect(ctx, dbURI)
			}
		}
	}
}

func (d *DB) reconnect(ctx context.Context, dbURI string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	newConn, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		slog.Error("failed to reconnect db", slog.Any("error", err))
		return
	}

	if newConn == nil {
		slog.Warn("new db connection is nil")
		return
	}

	if err := testConnection(ctx, newConn); err != nil {
		slog.Error("new db ping failed", slog.Any("error", err))
		newConn.Close()
		return
	}

	if d.conn != nil {
		d.conn.Close()
	}

	d.conn = newConn

	slog.Info("successfully reconnected to database")
}

func testConnection(ctx context.Context, db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return errors.Wrap(err, "db.Ping")
	}
	return nil
}

func applyMigrations(dbURI string) error {
	m, err := migrate.New("file://migrations/", dbURI)
	if err != nil {
		return errors.Wrap(err, "migrate.New")
	}

	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			slog.Error("failed to close migrate instance",
				slog.Any("sourceErr", sourceErr),
				slog.Any("dbErr", dbErr),
			)
		}
	}()

	err = m.Up()
	switch {
	case err == nil:
		slog.Info("migrations applied successfully")
	case errors.Is(err, migrate.ErrNoChange):
		slog.Info("no migrations to apply")
	default:
		return errors.Wrap(err, "m.Up")
	}

	return nil
}
func (d *DB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return d.conn.Query(ctx, query, args...)
}

func (d *DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return d.conn.QueryRow(ctx, query, args...)
}

func (d *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return d.conn.Exec(ctx, query, args...)
}

func (d *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return d.conn.Begin(ctx)
}
