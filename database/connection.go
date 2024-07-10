package database

import (
	"log"

	sqlx "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB = Connection()

func Connection() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "user=postgres password=123 dbname=db_doc sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}
	return db
}

func NewConnection() *sqlx.DB {
	otherDB, err := sqlx.Connect("postgres", "user=postgres password=123 dbname=db_aino_doc sslmode=disable")

	if err != nil {
		return nil
	}
	return otherDB
}
