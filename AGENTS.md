# AGENTS.md

## Cursor Cloud specific instructions

### Project Overview

Mantra API — a Go 1.26 backend with Gin (REST + GraphQL via gqlgen), GORM ORM, PostgreSQL, and JWT authentication. See the `justfile` for all standard dev commands (`just run`, `just test`, `just lint`, `just build`, etc.).

### Prerequisites

- **Go 1.26** must be installed at `/usr/local/go` (the VM's default Go 1.22 is too old).
- **PostgreSQL 16** must be running locally. The database `mantra_api` must exist with user `postgres` / password `postgres`.
- **just** command runner must be available in PATH.
- Copy `.env.example` to `.env` for local development (default values work out of the box).

### Starting PostgreSQL

```sh
sudo pg_ctlcluster 16 main start
```

### Running the App

```sh
just run   # starts on :8080
```

The app starts gracefully without PostgreSQL (DB features disabled), but full functionality requires a running database.

### Key Endpoints

| Endpoint | Description |
|---|---|
| `POST /api/v1/register` | User registration |
| `POST /api/v1/login` | User login |
| `POST /graphql` | GraphQL endpoint (mutations/queries) |
| `GET /graphql` | GraphQL Playground |
| `GET /swagger/*any` | Swagger UI |
| `GET /health` | Health check |

### Testing

Tests use SQLite in-memory (no PostgreSQL needed):

```sh
just test   # go test -v ./...
```

### Non-obvious Gotchas

- The `golangci-lint` binary is installed locally into `./bin/` by `just setup`, not globally. The `just lint` command references `./bin/golangci-lint`.
- The `.golangci.yml` uses **v2 format** (version field at top). Older golangci-lint versions will fail.
- GraphQL field names use `snake_case` (e.g., `product_name`, `created_at`), matching the Go GORM column tags.
- CGO is needed for SQLite tests (`gorm.io/driver/sqlite` uses `mattn/go-sqlite3`). The default `CGO_ENABLED=1` works with gcc installed.
