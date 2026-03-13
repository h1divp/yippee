package store

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
	"github.com/stephenafamo/bob"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Store struct {
	db bob.DB
}

// Path should be the root path
// Default: path = '$USER/.yippee'
// Also runs migrations that are embedded
func New(path string) (*Store, error) {
	sqlDB, err := sql.Open("sqlite", path+"/index.db")
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, sqlDB, embedMigrations)
	if err != nil {
		return nil, err
	}

	if _, err = provider.Up(context.Background()); err != nil {
		return nil, err
	}

	return &Store{db: bob.NewDB(sqlDB)}, nil
}

func (s *Store) Close() error {
	return s.db.DB.Close()
}

func (s *Store) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.DB.BeginTx(ctx, opts)
}

// executor returns a bob.Executor, preferring a transaction when one is provided.
func (s *Store) executor(tx *sql.Tx) bob.Executor {
	if tx != nil {
		return bob.NewTx(tx)
	}
	return s.db
}
