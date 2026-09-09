# Project & Coding Structure

> Living doc: describes current practice, not a mandate. Code and doc disagree? Trust the code, then fix the doc (or don't — nobody's blocked on it). Ask instead of guessing if something's unclear.

A directory-by-directory map. Read [Architecture](architecture.md) first if you haven't — this doc assumes you know the three-layer split it explains.

```
backend/
├── main.go                         entry point, delegates to cmd/
├── cmd/                            cobra CLI commands — wiring only, no business logic
│   ├── cmd_root.go                   root command, --env-file flag, config loading
│   ├── cmd_serve.go                  `serve`: builds every adapter/usecase/handler and starts the HTTP server
│   └── cmd_migrate.go                `migrate up|down|reset|create`: goose wrapper
│
├── internal/
│   ├── wiring/                      the composition root — Wire() builds every adapter/usecase/handler, called by both cmd_serve.go and tests/internal/apptest
│   │
│   ├── modules/                    the domain — one directory per bounded concern, see below
│   │   ├── user/                     account identity + credentials
│   │   ├── auth/                     session/token lifecycle (login, refresh, logout)
│   │   └── order/                    commission orders, hiring history
│   │                                 (snapshot, not exhaustive — more land as features ship; check the directory itself for the current list)
│   │
│   ├── adapters/                   implements each module's ports against real technology
│   │   ├── postgres/                 every module's repository, via bun — grouped by technology
│   │   └── token/                    JWT implementation of auth.TokenIssuer
│   │
│   ├── handler/rest/               HTTP entry point — huma operations, DTOs, auth/role middleware
│   │
│   └── pkg/                        cross-cutting, domain-agnostic — importable from anywhere
│       ├── config/                   YAML + env config loader
│       ├── database/                 Postgres/bun connection setup
│       ├── httpserver/               huma server, request-id/logging/recover/CORS middleware, /livez /readyz
│       ├── baserepo/                 generic CRUD + transaction plumbing shared by adapters
│       ├── apperror/                 the error taxonomy (see architecture.md)
│       ├── security/                 password hashing
│       ├── logger/                   slog setup
│       └── migrations/               embedded SQL migration files, applied via goose
│
└── docs/                           you are here
```

## Anatomy of one module: `internal/modules/user/`

Every module follows the same file layout — learn this once, every module reads the same way:

| File | Contents |
|---|---|
| `entity.go` | The plain domain struct (`User`) and any value types (`Role`). No `bun`/`json` tags — this type is shared by every layer, so it stays free of any one layer's concerns. |
| `port.go` | Ports the module offers (`UserUsecase`) and needs (`UserRepository`), plus command/result types those ports take (`RegisterInput`). |
| `usecase.go` | The struct implementing the driving port (`userUsecase`) and the business rules themselves. |
| `usecase_test.go` | Unit tests for the usecase against a hand-written in-memory fake of the repository — no database. |
| `errors.go` | Sentinel `apperror.Error` values the module returns (`ErrUserNotFound`, `ErrInvalidCredential`, ...). |

Other modules follow the same shape. `auth` does not depend on `user.UserUsecase`; it declares a smaller `UserIdentity` port (`Authenticate` + `GetByID`) that `user.UserUsecase` satisfies at wiring time, so auth reuses password-verification without seeing Register or reaching around into `user`'s repository.

## Anatomy of a Postgres adapter: `internal/adapters/postgres/user_repository.go`

Every file in this package follows the same shape:

1. An unexported bun row model (`userModel`) with `bun:"..."` tags — the *only* place those tags exist.
2. `newUserModel(*user.User) *userModel` and `(*userModel) toDomain() *user.User` — conversions at the boundary. The domain `User` struct never touches bun.
3. An unexported repository struct wrapping `baserepo.Executor`, with a compile-time assertion `var _ user.UserRepository = (*userRepository)(nil)`.
4. `NewUserRepository(db *bun.DB) user.UserRepository` — constructor returns the *port* type, not the concrete struct, so callers can't reach past the interface.

## Anatomy of `internal/handler/rest/`

Not one file per module like the other layers — instead:

| File | Contents |
|---|---|
| `<module>_handler.go` | One `*Handler` struct per module, each with a `Register(api huma.API)` method that declares every HTTP operation for that module (path, method, request/response DTOs). |
| `httperror.go` | `mapAppError` — the one place an `apperror.Code` becomes an HTTP status, shared by every handler. |
| `middleware.go` | `requireAuth`/`requireRole` — huma per-operation middleware, attached only to routes that need them. |
| `requestctx.go` | `AuthInfo` — what `requireAuth` injects into the request context, and how handlers read it back out. |

## Looking for X? It's in Y.

| You're looking for... | Look in... |
|---|---|
| A specific HTTP route/endpoint | `internal/handler/rest/<module>_handler.go`, its `Register` method |
| A business rule (e.g. "role must be customer or artist") | `internal/modules/<module>/usecase.go` |
| A SQL query / table shape | `internal/adapters/postgres/<module>_repository.go`; schema in `internal/pkg/migrations/*.sql` |
| JWT claims/signing | `internal/modules/auth/entity.go` (`TokenClaims`), `internal/adapters/token/jwt_issuer.go` |
| Config keys / env var overrides | `internal/pkg/config/config.yaml` (schema+defaults), `.env.example` (local overrides) |
| Request-id/logging/CORS/panic-recovery | `internal/pkg/httpserver/{middleware,cors}.go` |
| Auth/role guards on a specific route | `internal/handler/rest/middleware.go`, then that route's `huma.Operation.Middlewares` |
| App wiring / "how does it all connect" | `internal/wiring/wiring.go` — read top to bottom |
| Error codes / how a domain error becomes an HTTP status | `internal/pkg/apperror`, `mapAppError` in `internal/handler/rest/httperror.go` |
