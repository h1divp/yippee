package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

// Path should be the root path
// Default: path = '$USER/.yippee'
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"/index.db")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}
