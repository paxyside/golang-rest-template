package postgres

import (
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func applyMigrations(source, dbURI string) error {
	m, err := migrate.New(source, dbURI)
	if err != nil {
		return err
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil || dbErr != nil {
			slog.Error("failed to close migration", slog.Any("sourceErr", sourceErr), slog.Any("dbErr", dbErr))
		}
	}()

	err = m.Up()
	switch {
	case err == nil:
		slog.Info("migrations applied")
	case errors.Is(err, migrate.ErrNoChange):
		slog.Info("no new migrations to apply")
	default:
		return err
	}

	return nil
}
