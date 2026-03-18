package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/h1divp/yippee/internal/api"
	"github.com/h1divp/yippee/internal/config"
	"github.com/h1divp/yippee/internal/services"
	"github.com/h1divp/yippee/internal/store"
	"github.com/urfave/cli/v3"
)

func main() {

	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "serves the yippee fileserver, meant to be run as a background process or docker container",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					port := cmd.Uint16("port")

					basePath := config.Bootstrap()

					db, err := store.New(basePath)
					if err != nil {
						return fmt.Errorf("opening store: %w", err)
					}
					defer db.Close()

					sessionStore := store.NewSessionStore()
					defer sessionStore.Close()

					authServ := services.NewAuthService(db, sessionStore)
					router := api.New(authServ)

					addr := fmt.Sprintf(":%d", port)
					server := &http.Server{
						Addr:    addr,
						Handler: router.Handler(),
					}

					go func() {
						log.Printf("Serving on port %d", port)
						if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							log.Fatalf("server error: %v", err)
						}
					}()

					quit := make(chan os.Signal, 1)
					signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
					<-quit

					log.Println("Shutting down server...")
					if err := server.Shutdown(context.Background()); err != nil {
						return fmt.Errorf("shutting down server: %w", err)
					}

					return nil
				},
				Flags: []cli.Flag{
					&cli.Uint16Flag{
						Name:    "port",
						Value:   6767,
						Usage:   "port used by daemon",
						Aliases: []string{"p"},
					},
				},
			},
			{
				Name:        "init",
				Usage:       "initializes file structure",
				Description: "useful when manually setting up or debugging",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					basePath := config.Bootstrap()

					store, err := store.New(basePath)
					if err != nil {
						log.Fatal(err)
					}
					defer store.Close()
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
