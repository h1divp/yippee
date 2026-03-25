package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/h1divp/yippee/internal/dberrors"
	"github.com/h1divp/yippee/internal/models"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

func (s *Store) CreateUser(ctx context.Context, tx *sql.Tx, setter *models.UserSetter) (*models.User, error) {
	user, err := models.Users.Insert(setter).One(ctx, s.executor(tx))
	if err != nil {
		if errors.Is(dberrors.UserErrors.ErrUniqueSqliteAutoindexUsers1, err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := models.Users.Query(
		sm.Where(models.Users.Columns.Username.EQ(sqlite.Arg(username))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := models.FindUser(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
