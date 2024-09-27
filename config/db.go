package config

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func PSQLConn(env *Env) *sql.DB {
	var dbURI string
	if env.GOENV == "production" {
		dbURI = env.DB_URI
	} else {
		dbURI = env.DB_URI
	}
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		panic(err.Error())
	}
	return db
}
