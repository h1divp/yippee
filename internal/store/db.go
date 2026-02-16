package store

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Store struct {
	DB *sql.DB
}

// Path should be the root path
// Default: path = '$USER/.yippee'
// Also runs migrations that are embedded
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"/index.db")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}
