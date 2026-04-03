package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/h1divp/yippee/internal/models"
	"github.com/h1divp/yippee/internal/services"
)

var (
	ErrUnauthorized = errors.New("authentication required")
)

type contextKey string

const contextKeyUser contextKey = "user"

type Middleware struct {
	authService *services.AuthService
}

// NewMiddleware creates a Middleware with the required service dependencies.
func NewMiddleware(authService *services.AuthService) *Middleware {
	return &Middleware{authService: authService}
}

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.authService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth is like Auth but rejects the request with 401 if no valid
// session is present. Use this to wrap route groups that must be authenticated.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			writeErr(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		user, err := m.authService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserFromContext extracts the authenticated user from a request context.
// Returns nil if no user is present (unauthenticated request).
func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(contextKeyUser).(*models.User)
	return user
}

func writeErr(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
