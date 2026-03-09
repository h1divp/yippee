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
	store *store.Store
}

func NewAuthService(store *store.Store) *AuthService {
	return &AuthService{store}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*store.User, string, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("hashing password: %w", err)
	}

	user := store.User{
		Username:     username,
		PasswordHash: hash,
		Role:         "user",
	}

	id, err := s.store.CreateUser(ctx, nil, user)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, "", ErrUsernameTaken
		}
		return nil, "", fmt.Errorf("creating user: %w", err)
	}
	user.ID = id

	token, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("generating session token: %w", err)
	}

	sess := store.Session{
		Token:     token,
		UserID:    id,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if _, err = s.store.CreateSession(ctx, nil, sess); err != nil {
		return nil, "", fmt.Errorf("creating session: %w", err)
	}

	return &user, token, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*store.User, string, error) {
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

	sess := store.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if _, err = s.store.CreateSession(ctx, nil, sess); err != nil {
		return nil, "", fmt.Errorf("creating session: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (*store.User, error) {
	sess, err := s.store.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("fetching session: %w", err)
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
