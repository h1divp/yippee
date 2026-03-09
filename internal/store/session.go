package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/h1divp/yippee/internal/models"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

func (s *Store) CreateSession(ctx context.Context, tx *sql.Tx, setter *models.SessionSetter) (*models.Session, error) {
	sess, err := models.Sessions.Insert(setter).One(ctx, s.executor(tx))
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Store) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	sess, err := models.Sessions.Query(
		sm.Where(models.Sessions.Columns.Token.EQ(sqlite.Arg(token))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := models.Sessions.Delete(
		dm.Where(models.Sessions.Columns.Token.EQ(sqlite.Arg(token))),
	).Exec(ctx, s.db)
	return err
}
