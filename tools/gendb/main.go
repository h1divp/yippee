// gendb creates a temporary SQLite database (gen.db) with migrations applied.
// It is used as part of the code generation workflow: `go run ./tools/gendb`
// Intended to be run from the project root.
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func main() {
	// Remove stale gen.db if it exists so we start clean.
	_ = os.Remove("gen.db")

	db, err := sql.Open("sqlite", "gen.db")
	if err != nil {
		log.Fatalf("open gen.db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping gen.db: %v", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	if err := goose.Up(db, "internal/store/migrations"); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	log.Println("gen.db ready")
}
