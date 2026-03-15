package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/h1divp/yippee/internal/models"
	"github.com/h1divp/yippee/internal/store"
	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 2
	argon2SaltLen            = 16
	argon2KeyLen      uint32 = 32
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username is already taken")
)

type AuthService struct {
	store        *store.Store
	sessionStore *store.SessionStore
}

func NewAuthService(store *store.Store, sessionStore *store.SessionStore) *AuthService {
	return &AuthService{store, sessionStore}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("hashing password: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("generating session token: %w", err)
	}

	tx, err := s.store.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := s.store.CreateUser(ctx, tx, &models.UserSetter{
		Username:     omit.From(username),
		PasswordHash: omit.From(hash),
		Role:         omit.From("user"),
	})
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, "", ErrUsernameTaken
		}
		return nil, "", fmt.Errorf("creating user: %w", err)
	}

	s.sessionStore.Put(token, store.Session{
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // 1 month
	})

	if err = tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("committing transaction: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("fetching user: %w", err)
	}

	ok, err := verifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, "", fmt.Errorf("verifying password: %w", err)
	}
	if !ok {
		return nil, "", ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("generating session token: %w", err)
	}

	s.sessionStore.Put(token, store.Session{
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // 1 month
	})

	return user, token, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	sess, ok := s.sessionStore.Get(token)
	if !ok {
		return nil, ErrInvalidCredentials
	}

	if time.Now().After(sess.ExpiresAt) {
		return nil, ErrInvalidCredentials
	}

	user, err := s.store.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return user, nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLen)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return saltB64 + "$" + hashB64, nil
}

func verifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.SplitN(encodedHash, "$", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("decoding hash: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLen)

	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
