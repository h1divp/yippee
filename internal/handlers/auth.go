package handlers

import (
	"net/http"

	"github.com/h1divp/yippee/internal/services"
)

type AuthHandler struct {
	authServ *services.AuthService
}

func NewAuthHandler(authServ *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authServ,
	}
}

type RegisterBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement this handler and service

	// 1) Parses and validate body
	// 2) Checks code to see if valid
	// 3) Retrieves role for code
	// 4) Hash and salt password
	// 5) Attempt to create user
	// 6) On success, return http only, secure cookie
	// Cookie is valid for 30 days
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement this handler and service

	// 1) Parse login body and validate
	// 2) Attempt to sign in
	// 3) If success, return http only, secure cookie
	// Cookie is valid for 30 days
}

func (h *AuthHandler) SelfHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement this handler and services

	// Yes i know, probably better names for this handler
	// 1) Retrieve user from context
	// 2) Return the information
}
