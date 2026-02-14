package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	Bootstrap()
}

func Bootstrap() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not find home directory: %v", err)
	}

	basePath := filepath.Join(home, ".yippee")
	usersPath := filepath.Join(basePath, "users")
	thumbsPath := filepath.Join(basePath, "thumbs")

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		fmt.Printf("Initializing Yippee storage at: %s\n", basePath)
	}

	directories := []string{basePath, usersPath, thumbsPath}

	// 0755 is the standard for directories (drwxr-xr-x)
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Error creating directory %s: %v", dir, err)
		}
	}

	fmt.Println("Yippee! Finished scaffolding filesystem.")
}
