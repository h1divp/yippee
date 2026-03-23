package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Bootstraps filesystem structure
// Creates files and returns yippee root
func Bootstrap() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not find home directory: %v", err)
		return ""
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
			return ""
		}
	}

	fmt.Println("Yippee! Finished scaffolding filesystem.")
	return basePath
}

func ValidateStructure() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	basePath := filepath.Join(home, ".yippee")
	required := []string{
		basePath,
		filepath.Join(basePath, "users"),
		filepath.Join(basePath, "thumbs"),
	}

	for _, dir := range required {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return "", false
		}
	}

	return basePath, true
}
