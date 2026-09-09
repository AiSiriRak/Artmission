# Artmission Backend

## Prerequisites

- [Go 1.27+](https://go.dev/)
- [Docker + Docker Compose](https://www.docker.com/) (for local Postgres)
- [go-task](https://taskfile.dev/installation/) (optional — only needed for the `task` shortcuts below; the manual steps work without it)

## How to Run

### Option A: go-task (recommended)

Cross-platform, works on any OS:

```bash
task dev   # copies .env.example to .env if missing, starts Docker infra, and runs the API with hot reload
```

When you're done:

```bash
task down  # stops the Docker infra (Postgres)
```

### Option B: Manual

If you'd rather run each step yourself (or don't have go-task installed):

1. Start local infrastructure (Postgres):

```bash
docker compose up -d
```

2. Create a local environment file:

```bash
cp .env.example .env
```

The defaults in `.env.example` already match the compose stack — nothing to fill in for local dev.

3. Apply database migrations:

```bash
go run . --env-file .env migrate up
```

(The `serve` command below also runs pending migrations on startup, so this step is optional for local dev — it's here for when you need `migrate down`/`reset`/`create` on their own. See `task migrate-*`, e.g. `task migrate-create -- add_foo_column`.)

4. Start the API server:

```bash
go tool air                       # Run with hot reload
go run . --env-file .env serve    # Or without hot reload
```

5. To stop the infrastructure started manually:

```bash
docker compose down
```

### Once It's Running

Regardless of which option you used, the API is now available at:

```text
http://localhost:8080/api/v1
```

API Documentation (Swagger UI):

```text
http://localhost:8080/api/v1/docs
```

Health checks:

```bash
curl http://localhost:8080/livez   # process is up
curl http://localhost:8080/readyz  # process is up AND its dependencies (Postgres, ...) are reachable
```

### API Testing

`backend/bruno/` contains a [Bruno](https://www.usebruno.com/) collection for testing the API endpoints manually — open it in the Bruno app (or run it via `bru run`) to try requests against a running local server.

## Environment Variables

See `.env.example` — any nested key in `internal/pkg/config/config.yaml` can be overridden by an env var (e.g. `auth.jwt_secret` → `AUTH_JWT_SECRET`); the defaults already match the local `docker-compose.yaml` stack, so nothing is required for local dev beyond copying the example file.

## Documentation

New to the codebase? Start here:

- [Architecture](docs/architecture.md)
- [Project & Coding Structure](docs/project-structure.md)
- [Conventions](docs/conventions.md)
- [Adding a Feature](docs/adding-a-feature.md)
