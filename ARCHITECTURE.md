# Architecture

This document describes the architecture of **yippee!** — a self-hostable file server and personal cloud. It is intended for contributors who want to understand how the system fits together, and why it was built the way it was.

Go module: `github.com/h1divp/yippee`

## Table of Contents

- [High-Level Overview](#high-level-overview)
- [Entry Point and Bootstrap](#entry-point-and-bootstrap)
- [Code Generation Pipeline](#code-generation-pipeline)
- [Store Layer](#store-layer)
- [Services Layer](#services-layer)
- [Handlers Layer](#handlers-layer)
- [Generated Model API](#generated-model-api)
- [Database Schema](#database-schema)
- [Frontend](#frontend)
- [Key Design Decisions](#key-design-decisions)
- [Planned: File Storage](#planned-file-storage)
- [Current Status and Known Gaps](#current-status-and-known-gaps)

---

## High-Level Overview

```
SvelteKit SPA (web/)
        │ HTTP/JSON
   HTTP Handlers (internal/handlers)
        │
   Services (internal/services)     ← business logic
        │
   Store (internal/store)           ← data access, query builder
        │
   Models (internal/models)         ← GENERATED structs & query helpers
        │
   SQLite (modernc.org/sqlite)      ← pure Go, no CGo
```

The layering is strict and unidirectional. Handlers call services; services call the store; the store uses generated models to talk to SQLite. No layer calls upward. Business logic lives exclusively in services — handlers are thin request/response translators.

Runtime data lives in `~/.yippee/`:

| Path | Purpose |
|---|---|
| `~/.yippee/index.db` | SQLite database (metadata index only) |
| `~/.yippee/users/` | User-owned file trees |
| `~/.yippee/thumbs/` | Generated thumbnails, keyed by content hash |

**The local filesystem is the source of truth for file content.** SQLite holds only metadata. If `index.db` is deleted, no file data is lost — the index can be rebuilt by walking the filesystem.

---

## Entry Point and Bootstrap

**`cmd/yippee/main.go`** is the binary entrypoint. It calls `config.Bootstrap()`, opens the store, and defers close. The HTTP server is not yet wired.

**`internal/config`** contains `Bootstrap()`, which creates `~/.yippee/{users,thumbs}/` if they do not exist and returns the base path. Separating this from `main.go` keeps startup logic testable without running the full binary.

---

## Code Generation Pipeline

A significant portion of the codebase under `internal/` is generated from the database schema. This is a deliberate architectural choice: it eliminates hand-written SQL strings, prevents column name typos, and keeps struct definitions in sync with the schema automatically.

The generator is `bobgen-sqlite`, configured by `bobgen.yaml` at the project root. It reads a live SQLite database and emits three Go packages:

| Package | Contents | Use for |
|---|---|---|
| `internal/models` | Structs, setters, column refs, finders, relationship methods | All DB read/write operations |
| `internal/factory` | `New*` builder functions with random defaults | Writing tests without hand-rolling insert code |
| `internal/dberrors` | Typed constraint errors | Translating DB errors to domain errors without string matching |

**Never edit `internal/models`, `internal/factory`, or `internal/dberrors` by hand.** Changes are overwritten on the next `make generate`.

### The `make generate` pipeline

Running `make generate` does three things in sequence:

1. `tools/gendb/main.go` — opens a temporary `gen.db` and applies all migrations via Goose
2. `go tool bobgen-sqlite` — reads `gen.db`, generates Go code from the schema
3. Deletes `gen.db`

Generated files are committed to the repository. Generation never happens at runtime.

**Always run `make generate` after adding or modifying a migration.** If you add a new table, you must also add it to the `only:` block in `bobgen.yaml` before generating, or it will be silently excluded.

---

## Store Layer

`internal/store/` is the data access layer. It wraps a `bob.DB` and exposes typed methods that return generated model structs.

| File | Purpose |
|---|---|
| `db.go` | `Store` struct, `New()`, `Close()`, `BeginTx()`, `executor()` helper |
| `user.go` | `CreateUser`, `GetUserByUsername`, `GetUserByID` |
| `session.go` | `CreateSession`, `GetSessionByToken`, `DeleteSession` |
| `errors.go` | `ErrNotFound`, `ErrDuplicate` sentinel errors |
| `migrations/` | Goose-formatted SQL migration files |

### Transaction support

All write methods accept an optional `*sql.Tx` as their second parameter. Pass `nil` when no transaction is needed. Internally, the `executor(tx)` helper selects the transaction if one is provided, or falls back to the underlying DB connection. Store methods never need to branch manually on whether they are inside a transaction.

The store layer is transaction-aware but does not own transaction boundaries. Opening and committing transactions is the responsibility of the services layer.

### Error translation

The store translates low-level `dberrors.*` constraint errors into domain-level sentinels (`ErrNotFound`, `ErrDuplicate`). Callers above the store layer never need to inspect raw SQLite error codes.

---

## Services Layer

`internal/services/` contains all business logic. Currently it exposes `AuthService` with three methods: `Register`, `Login`, and `ValidateSession`.

### Auth details

- Passwords are hashed with argon2id (64 MB memory, 3 iterations, 2 parallelism) and stored as `{salt_b64}${hash_b64}`.
- Sessions use a 32-byte random token encoded as URL-safe base64, valid for 30 days.

### Transaction ownership

When multiple store writes must succeed or fail atomically, services open a transaction. The pattern is:

```go
tx, err := s.store.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("beginning transaction: %w", err)
}
defer tx.Rollback() // no-op after Commit

user, err := s.store.CreateUser(ctx, tx, setter)
if err != nil {
    return err
}

if err = tx.Commit(); err != nil {
    return fmt.Errorf("committing transaction: %w", err)
}
```

`defer tx.Rollback()` is always safe: if `Commit` has already been called, `Rollback` is a no-op.

### Error contracts

Services return domain-specific errors (`ErrInvalidCredentials`, `ErrUsernameTaken`), not HTTP status codes. Mapping errors to HTTP responses is the handler's job.

---

## Handlers Layer

`internal/handlers/` contains HTTP handlers and middleware. The standard library `net/http` is used directly — no third-party router.

| File | Purpose |
|---|---|
| `auth.go` | `AuthHandler` with `RegisterHandler`, `LoginHandler`, `SelfHandler` |
| `middleware.go` | `AuthMiddleware`, `UserFromContext` |
| `response.go` | `writeJSON`, `writeError`, `setSessionCookie` helpers |

### Middleware

`AuthMiddleware` reads the `session` cookie, validates it via `AuthService`, and injects a `*models.User` into the request context. If there is no cookie, the request passes through unauthenticated — individual handlers decide whether authentication is required.

Use `handlers.UserFromContext(ctx)` to extract the authenticated user from a handler's context.

### Session cookie

Name: `session`. Flags: `HttpOnly`, `Secure`, `SameSite=Lax`. Expiry: 30 days. Path: `/`.

### Handler responsibilities

Handlers are intentionally thin:

- Decode request bodies from JSON (`json.NewDecoder(r.Body)`)
- Validate required fields and return 400 if missing
- Call a service method
- Map service errors to HTTP status codes (e.g., `ErrUsernameTaken` → 409, `ErrInvalidCredentials` → 401)
- Write responses via `writeJSON` or `writeError`

Business logic does not belong in handlers.

---

## Generated Model API

The following examples illustrate how to use the generated types. These patterns appear throughout the store layer.

```go
// Find by primary key
user, err := models.FindUser(ctx, exec, id)

// Query with a WHERE clause — column names are type-safe, not strings
user, err := models.Users.Query(
    sm.Where(models.Users.Columns.Username.EQ(sqlite.Arg(username))),
).One(ctx, exec)

// Query multiple rows
sessions, err := models.Sessions.Query(
    sm.Where(models.Sessions.Columns.UserID.EQ(sqlite.Arg(userID))),
).All(ctx, exec)

// Insert — only set fields are included in the INSERT statement
user, err := models.Users.Insert(&models.UserSetter{
    Username:     omit.From("alice"),
    PasswordHash: omit.From(hash),
    Role:         omit.From("user"),
    // FullName and Email omitted → NULL in the database
}).One(ctx, exec)
// user.ID, user.CreatedAt, etc. are populated from the DB after insert

// Delete with a WHERE clause
_, err := models.Sessions.Delete(
    dm.Where(models.Sessions.Columns.Token.EQ(sqlite.Arg(token))),
).Exec(ctx, exec)

// Constraint errors — use errors.Is, never string matching
if errors.Is(err, dberrors.UserErrors.ErrUniqueSqliteAutoindexUsers1) {
    // username already taken
}
```

Nullable columns on generated structs use `null.Val[T]`, not `*T`. Setter fields use `omit.Val[T]`; use `omit.From(val)` to set a value and `omitnull.From(val)` for nullable columns that you want to set explicitly.

---

## Database Schema

Two tables exist today: `users` and `sessions`. All migrations live in `internal/store/migrations/` in Goose format with `-- +goose Up` and `-- +goose Down` blocks.

### Adding a new table

1. Run `goose create <name> sql` from `internal/store/`, then move the generated file into `migrations/`.
2. Write the SQL migration.
3. Add the table name to the `only:` block in `bobgen.yaml`.
4. Run `make generate`.

---

## Frontend

**Stack**: Svelte 5 (runes syntax), SvelteKit (SPA mode), TypeScript (strict), Tailwind CSS v4, daisyUI 5, pnpm.

SSR is disabled globally via `export const ssr = false` in `+layout.ts`. The application is a pure client-side SPA that communicates with the Go backend over HTTP/JSON.

### Routes

| Route | Description |
|---|---|
| `/login` | Login form |
| `/invite/[id]` | Invite acceptance (stub) |
| `/(protected)/files` | File browser root |
| `/(protected)/files/[...path]` | File browser at path |
| `/(protected)/s/[id]` | Shared file by ID |
| `/(protected)/shared` | Shared files list |

The `(protected)` layout group redirects to `/login` if `$user` is null, enforcing authentication at the route level.

The auth store (`src/lib/stores/auth.ts`) currently uses localStorage as a mock backend. Real API integration is not yet implemented.

---

## Key Design Decisions

**Pure Go, no CGo.** `modernc.org/sqlite` is used instead of `mattn/go-sqlite3`. This allows cross-compilation to any target platform without a C toolchain — `go build` just works.

**Generated models.** `bobgen-sqlite` generates type-safe Go code directly from the SQLite schema. This eliminates raw SQL strings, catches column name mistakes at compile time, and keeps struct definitions in sync with the schema without manual effort.

**Standard library HTTP.** `net/http` is used directly with no third-party router. This keeps the dependency tree minimal and avoids framework churn. The routing surface is small enough that a custom router adds no value.

**No globals; dependency injection via struct fields.** Every layer receives its dependencies through struct initialization. This makes the dependency graph explicit and unit testing straightforward.

**Filesystem-first, not database-first.** The intended design (not yet implemented) treats the filesystem as the authoritative source of truth. SQLite will hold only a derived metadata index. If `index.db` is lost, it can be rebuilt by walking the filesystem — no file data is ever stored exclusively in the database.

---

## Planned: File Storage

The core file serving functionality is not yet implemented. The intended approach is filesystem-first: files live under `~/.yippee/users/{username}/` on disk, and SQLite indexes their metadata. Users will be able to place files manually (via `cp`/`mv`) and have them detected automatically, without going through the API.

Change detection will use three layers: `fsnotify` for immediate events, a periodic background rescan as a safety net, and an on-demand directory diff when a user navigates to a folder. Thumbnails will be generated asynchronously and keyed by content hash under `~/.yippee/thumbs/`.

The planned tables are `files` (path, size, hash, mime type, mod time) and `shares` (public share links with optional expiry and download limits).

---

## Current Status and Known Gaps

- HTTP server routing is not yet wired in `cmd/yippee/main.go`.
- Invite code validation in `RegisterHandler` is a placeholder (`TODO`).
- Frontend auth is mocked against `localStorage` and is not connected to the real API.
- File upload, serving, browsing, sharing, and thumbnail generation are not yet implemented.
- `internal/auth/` is a legacy stub, superseded by `internal/services`. It can be deleted.
