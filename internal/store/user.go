package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/scan"
)

type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	FullName     *string   `db:"full_name"`
	Email        *string   `db:"email"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (s *Store) CreateUser(ctx context.Context, tx *sql.Tx, u User) (int64, error) {
	q := sqlite.Insert(
		im.Into("users", "username", "password_hash", "role", "full_name", "email"),
		im.Values(sqlite.Arg(u.Username, u.PasswordHash, u.Role, u.FullName, u.Email)),
	)
	res, err := bob.Exec(ctx, s.executor(tx), q)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrDuplicate
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	q := sqlite.Select(
		sm.Columns("id", "username", "password_hash", "role", "full_name", "email", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("username").EQ(sqlite.Arg(username))),
	)
	u, err := bob.One(ctx, s.db, q, scan.StructMapper[User]())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*User, error) {
	q := sqlite.Select(
		sm.Columns("id", "username", "password_hash", "role", "full_name", "email", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id))),
	)
	u, err := bob.One(ctx, s.db, q, scan.StructMapper[User]())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
