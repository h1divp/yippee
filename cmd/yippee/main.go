package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/h1divp/yippee/internal/config"
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
					fmt.Println("(NOT IMPLEMENTED) Serving on port:", port)
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
