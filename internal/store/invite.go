package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/h1divp/yippee/internal/models"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

func (s *Store) CreateInvite(ctx context.Context, tx *sql.Tx, setter *models.InviteSetter) (*models.Invite, error) {
	inv, err := models.Invites.Insert(setter).One(ctx, s.executor(tx))
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *Store) GetInviteByCode(ctx context.Context, code string) (*models.Invite, error) {
	inv, err := models.Invites.Query(
		sm.Where(models.Invites.Columns.Code.EQ(sqlite.Arg(code))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return inv, nil
}
