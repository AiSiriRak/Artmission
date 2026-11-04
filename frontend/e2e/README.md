# End-to-end tests (Playwright)

Black-box UI tests for the test plan in `docs/testplan/`. Each test title starts with its test case ID (`TC<epic>-<story>-<n> ...`), so `e2e-report/results.json` maps straight onto the coverage matrix.

Specs import `test` and `expect` from `support/test.ts`. After every test it saves the final page as `e2e-evidence/<test case ID>.png` (git-ignored); a case whose expected output appears before its last step also saves `e2e-evidence/<ID>-<label>.png`. Paste these into the test plan as the actual output. Cases that rely on a browser's own validation bubble (TC1-4-5) show nothing of it, because browser UI is not part of the page.

The tests drive the real frontend, API and database. Nothing is mocked.

## Run

The tests use their own database (`artmission_e2e`, in the same Postgres container) loaded with the `e2e` seed dataset, so manual testing and demos on the normal database never touch them. From `backend/`:

```bash
# terminal 1: docker infra, create + migrate + seed artmission_e2e, then the API on :8080
# (stop `task dev` first, both use :8080)
task e2e-serve

# terminal 2: reseed, then run the suite. Reseed before EVERY run: accepting an order,
# sending an extension request or submitting a draft uses up the seeded order it ran on.
task e2e-test

# or by hand, from frontend/
npm run e2e                                   # all
npx playwright test e2e/<name>.spec.ts        # one file
npx playwright show-report e2e-report/html    # HTML report of the last run
```

`next dev` on :3000 is started by Playwright unless it is already running. To go back to normal development, stop `task e2e-serve` and run `task dev`. Chromium comes from Playwright's own cache (`npx playwright install chromium` if it is missing). The Slip2Go settings are not needed: no test touches payment.

| Variable | Default |
|---|---|
| `E2E_BASE_URL` | `http://localhost:3000` |
| `E2E_API_BASE_URL` | `http://localhost:8080/api/v1` |

## How test data works

- **Accounts** are listed in `support/accounts.ts`. `auth.setup.ts` logs each one in through the login page once and saves the session under `e2e/.auth/` (git-ignored); specs pick one with `test.use({ storageState: authFile("artistBo") })`.
- **PENDING orders** are placed through the real API as the customer (`placeOrder` in `support/api.ts`), so they can be created as often as needed.
- **Upload files** the tests choose in the browser are in `e2e/fixtures/` (small valid JPEG, PNG and WebP). Invalid or oversized files are built inside the test.
- **Every other status** is seeded, because `IN_PROCESS` cannot be reached through the API without a real Slip2Go payment. The orders are in `backend/internal/pkg/seed/dataset_e2e.go`. Look one up by name with `findOrderId`; the names are part of the contract, so keep them stable. A test that changes an order needs one of its own; tests that only read or fail validation can share one.
- **Accounts a test changes** (register, edit, delete) are created with `createUser` and removed with `removeUser` (`support/api.ts`), so parallel tests never share them. The one exception is the seeded artist cleo, deleted by TC1-6-2 and restored by the next reseed.
