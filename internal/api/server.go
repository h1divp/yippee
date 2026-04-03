package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/h1divp/yippee/internal/handlers"
	"github.com/h1divp/yippee/internal/services"
)

type Server struct {
	handler http.Handler
}

func NewServer(authService *services.AuthService) *Server {
	authHandler := handlers.NewAuthHandler(authService)
	mw := NewMiddleware(authService)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /auth/register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /auth/login", authHandler.LoginHandler)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /auth/self", authHandler.SelfHandler)
	mux.Handle("/", mw.RequireAuth(protected))

	return &Server{handler: mux}
}

func (s *Server) Serve(port uint16) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Server listening on port %d\n", port)
	return srv.ListenAndServe()
}
