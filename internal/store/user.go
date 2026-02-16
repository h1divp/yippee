package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	FullName     *string
	Email        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Store) CreateUser(ctx context.Context, tx *sql.Tx, u User) (int64, error) {
	conn := pickConn(tx, s.DB)
	res, err := conn.ExecContext(
		ctx,
		`
		INSERT INTO users (
			username,
			password_hash,
			role,
			full_name,
			email
		) VALUES (?, ?, ?, ?, ?)
		`,
		u.Username,
		u.PasswordHash,
		u.Role,
		u.FullName,
		u.Email,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// TODO: Implement
func (s *Store) ListUsersByRole() ([]User, error) { return []User{}, fmt.Errorf("Not Implemented") }

// TODO: Implement
func (s *Store) GetUserById() error { return fmt.Errorf("Not Implemented") }
