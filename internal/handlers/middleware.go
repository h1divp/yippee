package handlers

import (
	"context"
	"net/http"

	"github.com/h1divp/yippee/internal/models"
	"github.com/h1divp/yippee/internal/services"
)

type contextKey string

const contextKeyUser contextKey = "user"

func AuthMiddleware(authServ *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := authServ.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext extracts the authenticated user from a request context.
// Returns nil if no user is present (unauthenticated request).
func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(contextKeyUser).(*models.User)
	return user
}
