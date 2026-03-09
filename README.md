# yippee!

A self-hostable file server and personal cloud.

yippee! is inspired by Copyparty and Google Drive. It ships as a single Go binary with no external runtime dependencies — SQLite stores metadata, the local filesystem is the source of truth. Early development; not production ready.

Data is stored in `~/.yippee/` (`users/`, `thumbs/`, `index.db`).

## Requirements

- Go 1.21+
- Node.js + pnpm (frontend development only)

## Quick Start

```bash
git clone https://github.com/h1divp/yippee
cd yippee

# Run the dev server
go run ./cmd/yippee

# Or build a binary
go build ./cmd/yippee
make build
```

**Frontend (from `web/`):**

```bash
pnpm dev    # dev server
pnpm build  # production build
```

**Tests:**

```bash
go test ./...
go test -race ./...
```

## Status

Early development. Core authentication (registration, login, sessions) is implemented. File upload, browsing, sharing, and thumbnails are not yet implemented. The HTTP router is not fully wired.

| Feature | Status |
|---|---|
| User auth (register/login) | Done |
| Session management | Done |
| File upload / serving | Not started |
| File browsing | Not started |
| Sharing | Not started |
| Thumbnails | Not started |

## Key Dependencies

| Package | Purpose |
|---|---|
| `modernc.org/sqlite` | Pure-Go SQLite (no CGo) |
| `github.com/stephenafamo/bob` | Type-safe SQL query builder |
| `github.com/pressly/goose/v3` | Database migrations |
| Svelte 5, SvelteKit | Frontend framework |
| Tailwind CSS v4, daisyUI 5 | Frontend styling |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for a detailed technical overview of the layer design, code generation, and conventions.

## License

MIT License. See [LICENSE](LICENSE).
