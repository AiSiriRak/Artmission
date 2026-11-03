import { test as setup } from "@playwright/test";
import { accounts, authFile, type AccountKey } from "./support/accounts";

// Logs in once per account through the real login page and stores the browser
// session (access token in localStorage + refresh cookie) for the specs to reuse.
for (const key of Object.keys(accounts) as AccountKey[]) {
  setup(`log in as ${key}`, async ({ page }) => {
    await page.goto("/login");
    await page.getByLabel("Email").fill(accounts[key].email);
    await page.getByPlaceholder("Enter your password").fill(accounts[key].password);
    await page.getByRole("button", { name: "Log in" }).click();
    // Where a user lands after logging in depends on the role (artists end up on
    // /orderhistory), so only wait until the login page has been left.
    await page.waitForURL((url) => url.pathname !== "/login");
    await page.context().storageState({ path: authFile(key) });
  });
}
