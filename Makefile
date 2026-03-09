.PHONY: generate build test

# Re-generate database models from the current schema.
# Requires the schema to be fully applied (via migrations) before running.
generate:
	@echo "==> Creating temp database with migrations..."
	go run ./tools/gendb
	@echo "==> Running bobgen-sqlite..."
	go tool bobgen-sqlite
	@rm -f gen.db
	@echo "==> Done. Generated files are in internal/models/"

build:
	go build ./cmd/yippee

test:
	go test ./...
