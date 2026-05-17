package migrator

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"strings"
)

func InitSQLMigrate(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		var dirtyErr migrate.ErrDirty
		if errors.As(err, &dirtyErr) {
			return fmt.Errorf(
				"database is dirty at migration version %d; fix schema_migrations and run force manually before restart: %w",
				dirtyErr.Version,
				err,
			)
		}

		if strings.Contains(err.Error(), "no migration found for version 0") {
			return fmt.Errorf(
				"schema_migrations points to version 0, but migration files start from version 1; fix schema_migrations manually and rerun migrations: %w",
				err,
			)
		}

		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
