# End-to-end tests (Playwright)

Black-box UI tests for the test plan in `docs/testplan/`. Each test title starts with its test case ID (`TC<epic>-<story>-<n> ...`), so `e2e-report/results.json` maps straight onto the coverage matrix.

The tests drive the real frontend, API and database. Nothing is mocked.

## Run

Everything below runs from the repo root. `task` can replace the `go run` lines (`task dev`, `task seed`).

```bash
# 1. infra (Postgres + RustFS)
cd backend && docker compose up -d --wait

# 2. schema and fixture data. Reseed before EVERY run: accepting an order, sending an
#    extension request or submitting a draft uses up the seeded order it ran on.
go run . --env-file .env migrate up
go run . --env-file .env seed reseed

# 3. API on :8080 (own terminal)
go run . --env-file .env serve

# 4. tests (starts `next dev` on :3000 itself unless it is already running)
cd ../frontend
npm run e2e                                   # all
npx playwright test e2e/smoke.spec.ts         # one file
npx playwright show-report e2e-report/html    # HTML report of the last run
```

Chromium comes from Playwright's own cache (`npx playwright install chromium` if it is missing). The Slip2Go settings are not needed: no test touches payment.

| Variable | Default |
|---|---|
| `E2E_BASE_URL` | `http://localhost:3000` |
| `E2E_API_BASE_URL` | `http://localhost:8080/api/v1` |

## How test data works

- **Accounts**: `jane@artmission.test` (customer), `ann@artmission.test` and `bo@artmission.test` (artists), all with `Password123!` (`support/accounts.ts`). `auth.setup.ts` logs each one in through the login page once and saves the session under `e2e/.auth/` (git-ignored); specs pick one with `test.use({ storageState: authFile("artistBo") })`.
- **PENDING orders** are placed through the real API as the customer (`placeOrder` in `support/api.ts`), so they need no seed and can be created as often as needed.
- **IN_PROCESS orders** cannot be reached through the API without a real Slip2Go payment, so the backend seed provides them: orders named `E2E ...` in `backend/internal/pkg/seed/orders_seed.go`, owned by `bo@artmission.test`, deadline today + 14 days. Look one up by name with `findOrderId`. A test that changes an order needs one of its own; tests that only read or fail validation can share one.

| Seeded order name | State |
|---|---|
| `E2E postpone - send request` | `IN_PROCESS`, no extension request |
| `E2E postpone - validation` | `IN_PROCESS`, no extension request; for tests that never succeed in sending one |
| `E2E postpone - request pending` | `IN_PROCESS`, with an extension request awaiting the customer |
| `E2E draft - JPEG`, `E2E draft - PNG`, `E2E draft - WebP` | `IN_PROCESS`, no drafts; one per file type that is submitted successfully |
| `E2E draft - validation` | `IN_PROCESS`, no drafts; for tests that never succeed in submitting one |
| `E2E cancelled order` | `CANCEL` (bo) |
| `Anime avatar` | `NOT_PAID` (ann), from the standard seed |
| `Dream landscape` | `IN_PROCESS` with a draft awaiting review (bo), from the standard seed |
| `Living room wall art` | `SUCCESS` (bo), from the standard seed |

`fixtures.spec.ts` is a harness check, not a test case: it fails fast if the seed is missing or stale.
