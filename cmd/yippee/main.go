package main

import (
	"log"

	"github.com/h1divp/yippee/internal/config"
	"github.com/h1divp/yippee/internal/store"
)

func main() {
	basePath := config.Bootstrap()

	store, err := store.New(basePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.DB.Close()
}
