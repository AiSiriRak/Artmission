# Conventions

> Living doc: describes current practice, not a mandate. Code and doc disagree? Trust the code, then fix the doc (or don't — nobody's blocked on it). Ask instead of guessing if something's unclear.

Patterns already in use across the codebase. New code should match these, not introduce a second way of doing the same thing.

## Naming

- **Interface = exported noun, no `I`/`Impl` prefix or suffix.** `UserUsecase`, `UserRepository`, `TokenIssuer`.
- **Struct implementing it = unexported, same name lowercased.** `userUsecase` implements `UserUsecase`; `userRepository` implements `UserRepository`.
- **Constructor = `New<InterfaceName>`, returns the interface type, not the struct.** `func NewUserUsecase(repo UserRepository) UserUsecase`. Callers only ever see the port — they can't reach past it to unexported fields.
- **Compile-time interface assertions on adapters**: `var _ user.UserRepository = (*userRepository)(nil)` right after the struct definition. Catches a missing method at build time with a clear error, instead of a confusing "does not implement" error somewhere else.

## Error handling

Domain code (`internal/modules/**`) never returns a raw `errors.New(...)` for anything a caller needs to branch on or that should reach the client — it returns an `*apperror.Error` (`internal/pkg/apperror`), built from one of six codes: `CodeInvalidInput`, `CodeUnauthorized`, `CodeForbidden`, `CodeNotFound`, `CodeConflict`, `CodeInternal`.

Two patterns, pick based on whether the error is a fixed, reusable case or has per-call detail:

- **Sentinel value**, in the module's `errors.go`, for errors any caller might compare against: `var ErrUserNotFound = apperror.NotFound("user not found")`.
- **Constructed inline**, for errors carrying call-specific detail or a wrapped cause: `apperror.Internal("failed to hash password", err)`.

The presentation layer (`internal/handler/rest`) is the *only* place a `Code` becomes an HTTP status — see `mapAppError` in `httperror.go`. If you're tempted to `import "net/http"` inside `internal/modules` or `internal/adapters`, that's the signal you're doing this in the wrong layer.

Unknown/unexpected errors (a plain `error` that isn't an `*apperror.Error`) should still resolve to a safe `500` at the handler boundary — never leak an unrecognized error's raw message to the client.

## Testing

- **Usecases get unit tests with hand-written in-memory fakes** — `internal/modules/<module>/usecase_test.go`, faking the module's own port interfaces (e.g. `fakeRepo` implementing `UserRepository` with a `map[string]*User`). No mocking framework, no code generation: a fake is a ~15-line struct. Fast (milliseconds), no Docker, runs in every `go test ./...`.
- **HTTP-level behavior gets a Cucumber/godog BDD suite** — real Postgres, real handlers, no mocks. See [Testing](testing.md).
- Table-driven tests where the cases are genuinely parallel (see `usecase_test.go` files for the shape); don't force a single "happy path only" test into a table just for consistency.

## Postgres adapters (bun)

- One unexported `bun:`-tagged row model per table, named `<entity>Model` (`userModel`), living only inside `internal/adapters/postgres`. The domain entity (`user.User`) never carries a bun tag.
- Conversion functions at the boundary: `new<Entity>Model(*domain.Entity) *<entity>Model` going in, `(*<entity>Model) toDomain() *domain.Entity` coming out.
- Repositories depend on `baserepo.Executor`, never `*bun.DB` directly — this is what lets a repository's queries transparently join an ambient transaction (via `baserepo.Transactioner`) without threading a `tx` parameter through every method.
- Generic CRUD (`Create`/`FindByID`/`ExistsByID`/`UpdateByID`/`DeleteByID`) is provided by `baserepo.BaseRepo[M]` — don't hand-write it again; embed/wrap it and add only the domain-specific queries (lookups by other columns, listings) alongside it.

## REST handlers (huma)

- Register each route with the verb-specific helper (`huma.Get`/`huma.Post`/`huma.Put`/...), not the lower-level `huma.Register{Operation}` — the HTTP method lives in the call, not a string field that can drift from it.
- `OperationID`: kebab-case, stable (drives generated client function names / OpenAPI `operationId`) — e.g. `view-hiring-history`.
- `Summary`: short PascalCase identifier matching the handler method it wraps, e.g. `ViewHiringHistory` for `h.viewHiringHistory`. This is what request-naming tools (Bruno, Postman, generated SDKs) use as the request's display name — keep it a name, not a sentence.
- `Description`: the human sentence (what used to live in `Summary`), e.g. `"View the authenticated customer's hiring history"`.
- `Tags`: one per module (`"auth"`, `"orders"`, ...) — groups operations in the generated docs UI and in codegen output.

## Config

- Schema and safe defaults live in one committed file: `internal/pkg/config/config.yaml`, embedded into the binary.
- Every nested YAML key is overridable by an environment variable using its dotted path uppercased with underscores: `auth.jwt_secret` → `AUTH_JWT_SECRET`. Secrets and per-environment values go here, never in `config.yaml`.
- Access config through the `Config` interface's accessor methods (`cfg.App()`, `cfg.Database()`, `cfg.Auth()`), not by reaching into a struct field directly — `cmd/cmd_serve.go` is the reference for how every existing command reads config.

## Wiring

All object construction happens by hand in one place, `internal/wiring.Wire` (`internal/wiring/wiring.go`), in dependency order: adapters → usecases → handlers → server. Both `cmd/cmd_serve.go` (the real server) and `tests/internal/apptest` (the BDD suite's in-process server) call it, so their object graphs can never diverge. No DI container, no `init()` magic. If you add a module, add its wiring here in the same order as everything else — see [Adding a Feature](adding-a-feature.md).
