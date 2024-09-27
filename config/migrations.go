package config

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func PSQLMigrate(db *sql.DB, env *Env) (*migrate.Migrate, error) {
	m, err := migrate.New("file://migrations/sql/", env.DB_URI)
	if err != nil {
		return nil, err
	}
	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			return nil, err
		}
	}
	return m, nil
}
