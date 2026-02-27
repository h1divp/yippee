package services

import "github.com/h1divp/yippee/internal/store"

type AuthService struct {
	store *store.Store
}

func NewAuthService(store *store.Store) *AuthService {
	return &AuthService{
		store,
	}
}
