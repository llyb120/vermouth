package vermouth

import "database/sql"

var db *sql.DB

func SetDB(_db *sql.DB) {
	db = _db
}

func GetDB() *sql.DB {
	return db
}