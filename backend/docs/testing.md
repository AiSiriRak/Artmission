# Testing

> Living doc: describes current practice, not a mandate. Code and doc disagree? Trust the code, then fix the doc (or don't — nobody's blocked on it). Ask instead of guessing if something's unclear.

Two tiers, each catching a different class of bug.

## Unit tests: business rules, no infrastructure

`internal/modules/<module>/usecase_test.go` — a usecase against a hand-written in-memory fake of its own port. No Docker, runs in every `go test ./...`. This doc is about the other tier.

## The BDD suite: `tests/`

One shared harness, one directory per domain — `auth/` shown below as the shape every other domain (`orders/`, ...) copies; this tree is illustrative, not a live listing, so it isn't kept in sync as domains are added:

```
tests/
├── internal/apptest/      shared bootstrap: container, app wiring, HTTP client, fixtures
└── auth/
    ├── features/*.feature   Gherkin scenarios
    ├── auth_test.go         starts Postgres + the app, runs the godog suite
    ├── steps_test.go        step definitions + the per-scenario context struct
    └── helpers_test.go      request payload builders
```

`tests/` is part of the same Go module (no nested `go.mod`) specifically so it can import `internal/...` and build the real app in-process — not a black-box client hitting a separately-deployed binary.

## Real Postgres, never the dev database

Each run starts one `testcontainers-go` Postgres container — once per test binary, not per scenario — applies every goose migration from `internal/pkg/migrations`, then wires the real app via `internal/wiring.Wire` (the same composition root `cmd_serve.go` calls) against it and serves it in-process via `httptest.Server`. It never touches `docker-compose.yaml`'s `artmission-db`; the container is disposed when the run ends.

## Isolation: unique data, not a reset between scenarios

Every scenario registers its own user with a unique username/email (`apptest.UniqueSuffix`, an `atomic.Uint64` counter — deterministic, not a probabilistic hash) instead of truncating tables between scenarios. One container serves every scenario in the binary; they never collide because none of them ever look at a row but their own.

## Opt-in via build tag, not part of `go test ./...`

Every file under `tests/` (including `apptest`) carries `//go:build integration`, so the default `go test ./...` never sees them and never needs Docker. Run the suite explicitly:

```bash
task test-bdd   # go test -tags=integration ./tests/... -v
```

## Writing feature files

- **Voice: `the user`, not `I`.** Multiple real roles exist (customer, artist); third person stays unambiguous once a scenario needs to name more than one.
- **`Given` is state only** — something already true before the scenario starts (`the user has a registered account`). The user *doing* something belongs in `When`, never `Given`.
- **`Then` describes the observable outcome** (`the system creates the account`), not the HTTP mechanics that prove it (status code, cookie, header) — those assertions live inside the step definition, not the Gherkin text.
- **`Scenario Outline` + `Examples`** for "same behavior, different invalid input," instead of one vague scenario or a pile of near-duplicates.

## Adding another domain's suite

Copy `tests/auth/`'s shape into `tests/<domain>/`, importing `tests/internal/apptest` for the container/app/client bootstrap rather than duplicating it. `tests/auth/auth_test.go` is the reference for the startup wiring.

The step definitions are mechanical once the `.feature` file is right — worth handing to an AI agent rather than hand-writing. Where your judgment is actually needed is upstream of that: getting the scenarios themselves declarative and correctly scoped (see "Writing feature files" above). Design and review the `.feature` file yourself, point the agent at it plus `tests/auth/` as the reference shape, and let it produce `steps_test.go`/`helpers_test.go` — then read the diff, don't just run it green.
