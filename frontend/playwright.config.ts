import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.E2E_BASE_URL ?? "http://localhost:3000";

// Black-box UI tests. They expect the whole stack to be running already
// (docker infra + Go API on :8080 + seeded database); see e2e/README.md.
export default defineConfig({
  testDir: "./e2e",
  outputDir: "e2e-results",
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [
    ["list"],
    ["html", { open: "never", outputFolder: "e2e-report/html" }],
    // Test titles start with the test case ID from the test plan (TC<epic>-<story>-<n>), so this file feeds the coverage matrix.
    ["json", { outputFile: "e2e-report/results.json" }],
  ],
  use: {
    baseURL,
    locale: "en-GB",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "setup", testMatch: /.*\.setup\.ts$/ },
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
      dependencies: ["setup"],
      testIgnore: /.*\.setup\.ts$/,
    },
  ],
  // Only the frontend is started on demand; the API and database are not.
  webServer: {
    command: "npm run dev",
    url: baseURL,
    reuseExistingServer: true,
    timeout: 120_000,
  },
});
