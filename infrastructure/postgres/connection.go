package postgres

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"emperror.dev/errors"

	"golang-template/internal/domain/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

type DB struct {
	mu sync.RWMutex

	uri           string
	pool          *pgxpool.Pool
	l             logger.Loggerer
	checkInterval time.Duration
	connTimeout   time.Duration
}

func Init(ctx context.Context, dbURI string, l logger.Loggerer) (*DB, error) {
	migrationsSource := viper.GetString("app.db.migrations_path")

	if err := applyMigrations(migrationsSource, dbURI); err != nil {
		return nil, errors.Wrap(err, "applyMigrations")
	}

	db := &DB{
		uri:           dbURI,
		l:             l,
		checkInterval: viper.GetDuration("app.db.healthcheck_interval"),
		connTimeout:   viper.GetDuration("app.db.connection_timeout"),
	}

	if err := db.connect(ctx); err != nil {
		return nil, errors.Wrap(err, "db.connect")
	}

	go db.healthChecker(ctx)

	return db, nil
}

func (db *DB) healthChecker(ctx context.Context) {
	ticker := time.NewTicker(db.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := db.Ping(ctx); err != nil {
				db.l.Error("database ping failed", slog.Any("error", err))

				if err = db.connect(ctx); err != nil {
					db.l.Error("database reconnect failed", slog.Any("error", err))
				}
			}
		}
	}
}

func (db *DB) connect(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	newPool, err := pgxpool.New(ctx, db.uri)
	if err != nil {
		return errors.Wrap(err, "pgxpool.New")
	}

	if err = newPool.Ping(ctx); err != nil {
		newPool.Close()
		return errors.Wrap(err, "newPool.Ping")
	}

	if db.pool != nil {
		db.pool.Close()
	}

	db.pool = newPool

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, db.connTimeout)
	defer cancel()

	if db.pool == nil {
		return errors.New("pool is nil")
	}

	return db.pool.Ping(ctx)
}

func (db *DB) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.pool != nil {
		db.pool.Close()
		db.pool = nil
	}
}
