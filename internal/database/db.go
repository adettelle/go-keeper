// Package database provides functionality for connecting to a PostgreSQL database
// using the `pgx` driver.
package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dbParams string) (*sql.DB, error) {
	log.Println("Connecting to DB ", dbParams)
	db, err := sql.Open("pgx", dbParams)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to DB ")
	return db, nil
}
