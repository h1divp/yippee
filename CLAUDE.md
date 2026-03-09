# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

**yippee!** is a self-hostable file server and personal cloud (inspired by Copyparty and Google Drive). Single Go binary, zero external dependencies, SQLite for metadata, local filesystem as source of truth. Early development.

Data lives in `~/.yippee/` — subdirs `users/`, `thumbs/`, and `index.db` (SQLite).

## Commands

### Go Backend

```bash
go run ./cmd/yippee        # Run the server
go build ./cmd/yippee      # Build binary
go test ./...              # Run all tests
go test ./internal/store   # Run tests for a specific package
```

### Frontend (run from `web/`)

```bash
pnpm dev       # Dev server
pnpm build     # Production build
pnpm check     # Type-check with svelte-check
pnpm lint      # Check formatting and lint
pnpm format    # Auto-fix formatting
```

## Architecture

### Layer Overview

```
SvelteKit SPA (web/)
        │ HTTP/JSON
   HTTP Handlers (internal/handlers)
        │
   Services (internal/services)     ← business logic
        │
   Store (internal/store)           ← data access, query builder
        │
   SQLite (modernc.org/sqlite)      ← pure Go, no CGo
```

### Go Backend (`cmd/`, `internal/`)

- **`cmd/yippee/main.go`** — Entrypoint. Calls `config.Bootstrap()`, opens the store, defers close. HTTP server not yet wired up.
- **`internal/config`** — `Bootstrap()` creates `~/.yippee/{users,thumbs}/` and returns the base path.
- **`internal/store`** — Data access layer. `Store` wraps `bob.DB` (query builder). SQLite opened via `modernc.org/sqlite`. Migrations embedded with `//go:embed` and run automatically via goose on startup.
  - `db.go` — Store struct, `New()`, `Close()`, `executor()` helper (prefers `*sql.Tx` over db).
  - `user.go` — `User` model + `CreateUser`, `GetUserByUsername`, `GetUserByID`.
  - `session.go` — `Session` model + `CreateSession`, `GetSessionByToken`, `DeleteSession`.
  - `errors.go` — `ErrNotFound`, `ErrDuplicate` sentinel errors.
  - `migrations/` — Goose-formatted SQL files (up/down blocks).
- **`internal/services`** — Business logic. `AuthService` implements `Register`, `Login`, `ValidateSession`.
  - Passwords hashed with argon2id (64 MB memory, 3 iterations, 2 parallelism). Stored as `{salt_b64}${hash_b64}`.
  - Sessions use 32-byte random token (URL-safe base64), valid 30 days.
- **`internal/handlers`** — HTTP handlers and middleware.
  - `auth.go` — `AuthHandler` with `RegisterHandler`, `LoginHandler`, `SelfHandler`.
  - `middleware.go` — `AuthMiddleware` reads `session` cookie, validates via `AuthService`, injects `*store.User` into context. Passes through if no cookie (handlers decide whether auth is required).
  - `response.go` — `writeJSON`, `writeError`, `setSessionCookie` helpers.
- **`internal/auth`** — Legacy stub, superseded by `internal/services`. Safe to remove.

### Database

Two tables: `users` and `sessions`. Migrations in `internal/store/migrations/`.

To create a new migration: run `goose create <name> sql` from `internal/store/`, then move the file into `migrations/`. Follow the `<timestamp>_<description>.sql` naming pattern.

Session cookie: name `session`, `HttpOnly`, `Secure`, `SameSite=Lax`, 30-day expiry, path `/`.

### SvelteKit Frontend (`web/`)

- **Stack**: Svelte 5 (runes), SvelteKit, TypeScript (strict), Tailwind CSS v4, daisyUI 5, pnpm.
- **SSR disabled** globally (`export const ssr = false` in `+layout.ts`) — pure SPA.
- **Auth store** (`src/lib/stores/auth.ts`): Mock `login`/`logout`/`checkSession` backed by `localStorage`. Real API integration not yet implemented.
- **Route protection**: `(protected)` layout group redirects to `/login` if `$user` is null.
- **Routes**:
  - `/login` — Login form
  - `/invite/[id]` — Invite acceptance (stub)
  - `/(protected)/files` and `/(protected)/files/[...path]` — File browser
  - `/(protected)/s/[id]` — Shared file by ID
  - `/(protected)/shared` — Shared files list

## Key Dependencies

### Go

| Package | Purpose |
|---|---|
| `modernc.org/sqlite` | Pure-Go SQLite driver (no CGo) |
| `github.com/pressly/goose/v3` | Database migrations |
| `github.com/stephenafamo/bob` | Type-safe SQL query builder |
| `github.com/stephenafamo/scan` | Struct scanning from `sql.Rows` |
| `golang.org/x/crypto` | Argon2id password hashing |

### Frontend

| Package | Purpose |
|---|---|
| `svelte` (v5) | UI framework |
| `@sveltejs/kit` | App framework (SPA mode) |
| `tailwindcss` (v4) | Utility-first CSS |
| `daisyui` (v5) | Component library |

## Conventions and Best Practices

### Go — General

- **Standard library HTTP**: Use `net/http` directly — no third-party routers. Handlers are `http.HandlerFunc` or methods on handler structs.
- **Error handling**: Return errors up the call stack. Use sentinel errors (`ErrNotFound`, `ErrDuplicate`, `ErrInvalidCredentials`) for expected failure cases. Check with `errors.Is()`. Wrap unexpected errors with `fmt.Errorf("context: %w", err)`.
- **Context propagation**: Pass `context.Context` as the first argument to all store and service methods. Handlers get context from `r.Context()`.
- **No globals**: Dependencies are injected via struct fields (e.g., `AuthService` holds `*store.Store`, `AuthHandler` holds `*services.AuthService`).

### Go — Store Layer

- All store methods take `ctx context.Context` as the first parameter.
- Methods that write data accept an optional `*sql.Tx` parameter to support transactions. Use the `executor(tx)` helper to pick between the transaction and the default db connection.
- Use `bob` for building queries and `scan` for mapping rows to structs. Prefer these over raw `db.Query` strings.
- Models use `db:"column_name"` struct tags for scan mapping.
- Nullable fields use pointer types (e.g., `*string` for `full_name`, `email`).

### Go — Services Layer

- Services contain all business logic. Handlers should not contain business logic beyond request parsing and response writing.
- Services return domain-specific errors (e.g., `ErrInvalidCredentials`, `ErrUsernameTaken`), not HTTP status codes.
- Crypto operations (hashing, token generation) live in services, not handlers.

### Go — Handlers Layer

- Request bodies are decoded from JSON using `json.NewDecoder(r.Body)`.
- Input validation happens at the handler level (check required fields, return 400).
- Use `writeJSON(w, status, data)` and `writeError(w, status, msg)` for consistent JSON responses.
- Map service errors to HTTP status codes in handlers (e.g., `ErrUsernameTaken` → 409, `ErrInvalidCredentials` → 401).
- Middleware injects values into context; handlers extract with type assertions.

### Go — Testing

- Use the standard `testing` package. No test framework dependencies.
- Table-driven tests where applicable.
- Test files live alongside the code they test (`foo_test.go` next to `foo.go`).

### Frontend

- Svelte 5 runes syntax (not legacy `$:` reactivity).
- TypeScript strict mode — avoid `any`.
- Use `$lib/` alias for imports from `src/lib/`.
- daisyUI component classes for UI elements.
- Keep route pages thin — extract logic into stores or lib modules.

## Incomplete / TODO

- HTTP server and routing not yet wired in `main.go`.
- Invite code validation in `RegisterHandler` (placeholder TODO).
- Frontend auth is mocked — needs real API integration.
- File upload, serving, browsing, sharing, thumbnails — not yet implemented.
- `internal/auth/` legacy stub can be deleted.
