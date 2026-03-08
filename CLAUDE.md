# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**yippee!** is a self-hostable file server and personal cloud (inspired by Copyparty and Google Drive). It is designed as a single Go binary with zero external system dependencies, using SQLite for metadata and the local filesystem as the source of truth. The project is in early development.

Data is stored in `~/.yippee/` with subdirs `users/`, `thumbs/`, and `index.db` (SQLite).

## Commands

### Go Backend
```bash
go run ./cmd/yippee      # Run the server
go build ./cmd/yippee    # Build binary
go test ./...            # Run all tests
go test ./internal/store # Run tests for a specific package
```

### Frontend (run from `web/`)
```bash
pnpm dev          # Dev server
pnpm build        # Production build
pnpm check        # Type-check with svelte-check
pnpm lint         # Check formatting and lint
pnpm format       # Auto-fix formatting
```

## Architecture

### Go Backend (`cmd/`, `internal/`)

- **`cmd/yippee/main.go`**: Entrypoint — bootstraps filesystem, opens DB.
- **`internal/config`**: `Bootstrap()` creates the `~/.yippee/` directory tree and returns the base path.
- **`internal/store`**: SQLite connection via `modernc.org/sqlite` (pure Go, no CGo). Migrations are embedded at compile time using `//go:embed migrations/*.sql` and run automatically via [goose](https://github.com/pressly/goose) on startup. The `Store` struct wraps `*sql.DB` and exposes query methods. Use `pickConn(tx, db)` to support optional transactions in store methods.
- **`internal/auth`**: `AuthService` holds a reference to `*store.Store` and will implement login/register logic (currently stubs).

**Migrations** live in `internal/store/migrations/` as goose-formatted `.sql` files (up/down blocks). New migrations should follow the naming pattern `<timestamp>_<description>.sql`.

### SvelteKit Frontend (`web/`)

- **Stack**: Svelte 5, SvelteKit, TypeScript, Tailwind CSS v4, daisyUI, pnpm.
- **SSR**: Disabled globally (`export const ssr = false` in `src/routes/+layout.ts`) — this is a pure SPA.
- **Auth**: `src/lib/stores/auth.ts` — Svelte writable stores (`user`, `loading`) + mock `login`/`logout`/`checkSession` functions backed by `localStorage`. Real API integration is not yet implemented.
- **Route protection**: The `(protected)` route group's layout (`src/routes/(protected)/+layout.svelte`) redirects to `/login` if `$user` is null.
- **Routes**:
  - `/login` — public login page
  - `/(protected)/files` — file browser
  - `/(protected)/s/[id]` — shared file by ID
  - `/(protected)/shared` — shared files list
