package auth

import (
	"fmt"

	"github.com/h1divp/yippee/internal/store"
)

type AuthService struct {
	store *store.Store
}

func NewService(store *store.Store) *AuthService {
	return &AuthService{
		store,
	}
}

func (s *AuthService) Register() error {
	return fmt.Errorf("Not implemented yet.")
}

func (s *AuthService) Login() error {
	return fmt.Errorf("Not implemented yet.")
}
