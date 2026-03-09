# Contributing to yippee!

Thank you for your interest in contributing. This document covers everything you need to get started, follow project conventions, and submit changes.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Database Migrations](#database-migrations)
- [Code Conventions](#code-conventions)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

---

## Getting Started

**Prerequisites:**

- Go 1.21 or later
- Node.js (LTS recommended)
- pnpm

**Clone and set up:**

```bash
git clone https://github.com/h1divp/yippee.git
cd yippee
cd web && pnpm install
```

The server stores all data in `~/.yippee/` (`users/`, `thumbs/`, `index.db`). This directory is created automatically on first run.

---

## Development Workflow

### Backend

```bash
go run ./cmd/yippee   # Run the server
go build ./cmd/yippee # Build the binary
make build            # Build via Makefile
```

### Frontend

Run all frontend commands from the `web/` directory.

```bash
pnpm dev     # Start the development server
pnpm build   # Build for production
pnpm check   # Type-check with svelte-check
pnpm lint    # Check formatting and lint
pnpm format  # Auto-fix formatting issues
```

### Code Generation

```bash
make generate
```

Run `make generate` after every migration change. It performs three steps in sequence:

1. Runs `tools/gendb/main.go` — creates a temporary `gen.db` and applies all migrations.
2. Runs `go tool bobgen-sqlite` — reads `gen.db` and generates Go code from the schema.
3. Deletes `gen.db`.

Generated files are committed to the repository. Always commit updated generated files alongside the migration that required them.

---

## Project Structure

```
cmd/
  yippee/           Entry point. Calls config.Bootstrap(), opens the store.

internal/
  config/           Bootstrap() — creates ~/.yippee subdirectories, returns base path.
  store/            Data access layer. Store wraps bob.DB. Contains migrations/.
  services/         Business logic. AuthService: Register, Login, ValidateSession.
  handlers/         HTTP handlers and middleware. No business logic.
  models/           GENERATED. Structs, setters, column refs, finders. Do not edit.
  factory/          GENERATED. Test fixture builders for all tables.
  dberrors/         GENERATED. Typed constraint errors for use with errors.Is().

web/                SvelteKit SPA. SSR disabled (pure SPA mode).
tools/
  gendb/            Dev tool used by make generate. Not part of the server binary.
```

---

## Database Migrations

1. Create a new migration file:

   ```bash
   # Run from internal/store/
   goose create <name> sql
   ```

2. Move the generated file into `internal/store/migrations/`.

3. Write the migration using Goose format:

   ```sql
   -- +goose Up
   CREATE TABLE example ( ... );

   -- +goose Down
   DROP TABLE example;
   ```

4. Add the table name to the `only:` block in `bobgen.yaml`. New tables are excluded from code generation unless explicitly listed there.

5. Run `make generate` to regenerate models, factories, and error types.

6. Commit the migration file and all regenerated files together.

---

## Code Conventions

### Go — General

- Use `net/http` directly. Do not add third-party HTTP routers.
- Inject dependencies via struct fields. Do not use globals.
- Pass `context.Context` as the first argument to every store and service method. Never store a context in a struct field.
- Do not use `context.WithValue` to pass services, DB handles, or configuration — those belong in struct fields.
- Do not define interfaces preemptively. Define them at the point of use, only when there are at least two concrete implementations or a clear test-seam requirement.
- Avoid package names like `util`, `common`, `helpers`, or `misc`. Each package should have a single, clear responsibility.

### Go — Error Handling

- Return errors up the call stack. Do not swallow them.
- Wrap errors at each layer boundary: `fmt.Errorf("creating session: %w", err)`.
- Use sentinel errors (`ErrNotFound`, `ErrDuplicate`, `ErrInvalidCredentials`) for expected failure cases and check them with `errors.Is()`.
- Never assign an error to `_` without an explicit reason.
- Do not use `panic` for normal error flow. Reserve it for truly unrecoverable programmer errors at startup.

### Go — Layer Boundaries

- **Handlers**: parse requests, validate inputs, map service errors to HTTP status codes, write responses. No business logic.
- **Services**: all business logic, crypto operations, transaction ownership. Return domain errors, not HTTP status codes.
- **Store**: data access only. Translate DB-level errors (`dberrors.*`) into domain sentinel errors. Never own transaction boundaries.

### Go — Transactions

- Begin transactions in the services layer, never in handlers or the store.
- Always `defer tx.Rollback()` immediately after `BeginTx`. It is a no-op after `Commit()`.
- Pass `*sql.Tx` into store methods. Pass `nil` when no transaction is needed.

### Go — Generated Code

- Never manually edit files in `internal/models`, `internal/factory`, or `internal/dberrors`.
- Use generated column references in WHERE clauses: `models.Users.Columns.Username.EQ(sqlite.Arg(val))`. Do not use raw column name strings.
- Nullable columns on generated structs are `null.Val[T]`, not `*T`.
- Setter fields use `omit.Val[T]`. Set a value with `omit.From(val)`. Omit a field to leave it out of the INSERT entirely.

### Frontend

- Use Svelte 5 runes syntax. Do not use legacy `$:` reactive declarations.
- TypeScript strict mode is enforced. Do not use `any`.
- Use daisyUI component classes for UI elements.
- Keep route pages thin. Extract logic into stores or `src/lib/` modules.
- Import from `src/lib/` using the `$lib/` alias.

---

## Testing

```bash
go test ./...              # Run all tests
go test -race ./...        # Run all tests with the race detector
go test ./internal/store   # Run tests for a specific package
```

- Use the standard `testing` package. Do not add test framework dependencies.
- Write table-driven tests where applicable.
- Place test files alongside the code they test: `foo_test.go` next to `foo.go`.
- Use `internal/factory` (generated) to create test fixtures. Do not hand-roll insert code in tests.
- Run the race detector regularly, especially when adding concurrent code.

---

## Submitting Changes

1. Fork the repository and create a branch from `main`.
2. Name branches using the format: `feat/<name>`, `fix/<name>`, or `docs/<name>`.
3. Write commit messages in imperative present tense: "Add X", "Fix Y", "Update Z".
4. Run `go test -race ./...` and `pnpm lint` (from `web/`) before pushing.
5. If your change includes a migration, run `make generate` and commit the regenerated files.
6. Open a pull request against `main`. Include a description of what the change does and why.
7. Keep PRs small and focused. Address one concern per pull request.

---

## Extras

1. Please follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) when making commits.
2. When using AI coding agents, please please please check everything over. Run all tests, etc.
3. Remove all agentic collaborators from the commit history. We want to uplift our human contributors.
4. Please don't submit 100% agentic made PRs, we will not accept these if they are blatant.
5. Please leave `good first issue` issues to human contributors!

---

## Reporting Issues

Open a GitHub issue and include:

- What you expected to happen.
- What actually happened.
- Steps to reproduce the problem.
- Your Go version (`go version`) and operating system.
