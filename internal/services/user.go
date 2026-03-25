package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/h1divp/yippee/internal/models"
	"github.com/h1divp/yippee/internal/store"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func ParseRole(s string) (Role, error) {
	switch Role(s) {
	case RoleAdmin, RoleUser:
		return Role(s), nil
	default:
		return "", fmt.Errorf("invalid role %q: must be \"admin\" or \"user\"", s)
	}
}

type UserMeta struct {
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserService struct {
	store    *store.Store
	basePath string
}

func NewUserService(store *store.Store, basePath string) *UserService {
	return &UserService{store: store, basePath: basePath}
}

func (s *UserService) CreateUser(ctx context.Context, username, password string, role Role) (*models.User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, nil, &models.UserSetter{
		Username:     omit.From(username),
		PasswordHash: omit.From(hash),
		Role:         omit.From(string(role)),
	})
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}

	homeDir := filepath.Join(s.basePath, "users", username)
	if err := os.MkdirAll(homeDir, 0o700); err != nil {
		return nil, fmt.Errorf("creating user home directory: %w", err)
	}

	meta := UserMeta{
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		CreatedAt:    user.CreatedAt,
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling user meta: %w", err)
	}
	metaPath := filepath.Join(homeDir, ".meta")
	if err := os.WriteFile(metaPath, metaBytes, 0o600); err != nil {
		return nil, fmt.Errorf("writing user meta file: %w", err)
	}

	return user, nil
}
