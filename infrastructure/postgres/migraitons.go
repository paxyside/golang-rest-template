package postgres

import (
	"emperror.dev/errors"
	"github.com/golang-migrate/migrate/v4"

	// Import Postgres driver for migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// Import file source driver for migrate.
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
			return
		}
	}()

	err = m.Up()
	switch {
	case err == nil:
		return nil
	case errors.Is(err, migrate.ErrNoChange):
		return errors.Wrap(err, "migrate.ErrNoChange")
	default:
		return errors.Wrap(err, "migrate.Up")
	}
}
