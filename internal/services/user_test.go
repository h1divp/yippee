package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/h1divp/yippee/internal/services"
	"github.com/h1divp/yippee/internal/store"
)

// setupTestEnv creates a temp directory with the expected yippee structure
// and returns a store, basePath, and cleanup function.
func setupTestEnv(t *testing.T) (*store.Store, string) {
	t.Helper()
	basePath := t.TempDir()

	// Create the users subdirectory that config.Bootstrap normally creates
	if err := os.MkdirAll(filepath.Join(basePath, "users"), 0o755); err != nil {
		t.Fatalf("creating users dir: %v", err)
	}

	s, err := store.New(basePath)
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	return s, basePath
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    services.Role
		wantErr bool
	}{
		{"valid admin", "admin", services.RoleAdmin, false},
		{"valid user", "user", services.RoleUser, false},
		{"empty string", "", "", true},
		{"unknown role", "superadmin", "", true},
		{"wrong case", "Admin", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := services.ParseRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRole(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRole(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateUser_HappyPath(t *testing.T) {
	s, basePath := setupTestEnv(t)
	svc := services.NewUserService(s, basePath)

	user, err := svc.CreateUser(context.Background(), "alice", "password123", services.RoleUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Username != "alice" {
		t.Errorf("username = %q, want %q", user.Username, "alice")
	}
	if user.Role != "user" {
		t.Errorf("role = %q, want %q", user.Role, "user")
	}

	// Verify home directory was created
	homeDir := filepath.Join(basePath, "users", "alice")
	info, err := os.Stat(homeDir)
	if err != nil {
		t.Fatalf("home directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("home path is not a directory")
	}

	// Verify .meta file was created with correct contents
	metaBytes, err := os.ReadFile(filepath.Join(homeDir, ".meta"))
	if err != nil {
		t.Fatalf("reading .meta file: %v", err)
	}
	var meta services.UserMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		t.Fatalf("unmarshaling .meta file: %v", err)
	}
	if meta.Role != "user" {
		t.Errorf("meta role = %q, want %q", meta.Role, "user")
	}
	if meta.PasswordHash == "" {
		t.Error("meta password_hash is empty")
	}
	if meta.CreatedAt.IsZero() {
		t.Error("meta created_at is zero")
	}
}

func TestCreateUser_AdminRole(t *testing.T) {
	s, basePath := setupTestEnv(t)
	svc := services.NewUserService(s, basePath)

	user, err := svc.CreateUser(context.Background(), "bob", "password123", services.RoleAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Role != "admin" {
		t.Errorf("role = %q, want %q", user.Role, "admin")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	s, basePath := setupTestEnv(t)
	svc := services.NewUserService(s, basePath)

	_, err := svc.CreateUser(context.Background(), "alice", "password123", services.RoleUser)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = svc.CreateUser(context.Background(), "alice", "otherpass123", services.RoleUser)
	if !errors.Is(err, services.ErrUsernameTaken) {
		t.Errorf("error = %v, want ErrUsernameTaken", err)
	}
}
