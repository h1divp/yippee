package api

import (
	"net/http"

	"github.com/h1divp/yippee/internal/handlers"
	"github.com/h1divp/yippee/internal/services"
)

type Router struct {
	authHandler *handlers.AuthHandler
}

func New(authServ *services.AuthService) *Router {
	return &Router{
		authHandler: handlers.NewAuthHandler(authServ),
	}
}

func (r *Router) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", r.authHandler.RegisterHandler)
	mux.HandleFunc("POST /auth/login", r.authHandler.LoginHandler)
	mux.HandleFunc("GET /auth/self", r.authHandler.SelfHandler)

	return mux
}

func (r *Router) Handler() http.Handler {
	return handlers.AuthMiddleware(r.authHandler.AuthServ())(r.Routes())
}
