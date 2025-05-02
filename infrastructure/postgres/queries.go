package postgres

import (
	"context"
	"emperror.dev/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
)

func (db *DB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if err := db.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "db.Ping")
	}

	return db.pool.Query(ctx, query, args...)
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if err := db.Ping(ctx); err != nil {
		return nil
	}

	return db.pool.QueryRow(ctx, query, args...)
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if err := db.Ping(ctx); err != nil {
		return pgconn.CommandTag{}, errors.Wrap(err, "db.Ping")
	}

	return db.pool.Exec(ctx, query, args...)
}

func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if err := db.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "db.Ping")
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		db.l.Error("failed to begin transaction", slog.Any("error", err))
		return nil, errors.Wrap(err, "db.pool.Begin")
	}

	return tx, nil
}

func (db *DB) Commit(ctx context.Context, tx pgx.Tx) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if err := db.Ping(ctx); err != nil {
		return errors.Wrap(err, "db.Ping")
	}

	if err := tx.Commit(ctx); err != nil {
		db.l.Error("failed to commit transaction", slog.Any("error", err))
		return errors.Wrap(err, "tx.Commit")
	}

	return nil
}

func (db *DB) Rollback(ctx context.Context, tx pgx.Tx) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if err := db.Ping(ctx); err != nil {
		return errors.Wrap(err, "db.Ping")
	}

	if err := tx.Rollback(ctx); err != nil {
		db.l.Error("failed to rollback transaction", slog.Any("error", err))
		return errors.Wrap(err, "tx.Rollback")
	}

	return nil
}
