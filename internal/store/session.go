package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/scan"
)

type Session struct {
	ID        int64     `db:"id"`
	Token     string    `db:"token"`
	UserID    int64     `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *Store) CreateSession(ctx context.Context, tx *sql.Tx, sess Session) (int64, error) {
	q := sqlite.Insert(
		im.Into("sessions", "token", "user_id", "expires_at"),
		im.Values(sqlite.Arg(sess.Token, sess.UserID, sess.ExpiresAt)),
	)
	res, err := bob.Exec(ctx, s.executor(tx), q)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	q := sqlite.Select(
		sm.Columns("id", "token", "user_id", "expires_at", "created_at"),
		sm.From("sessions"),
		sm.Where(sqlite.Quote("token").EQ(sqlite.Arg(token))),
	)
	sess, err := bob.One(ctx, s.db, q, scan.StructMapper[Session]())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	q := sqlite.Delete(
		dm.From("sessions"),
		dm.Where(sqlite.Quote("token").EQ(sqlite.Arg(token))),
	)
	_, err := bob.Exec(ctx, s.db, q)
	return err
}
