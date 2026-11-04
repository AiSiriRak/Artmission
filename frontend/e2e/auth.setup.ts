import { test as setup } from "@playwright/test";
import { accounts, authFile, type AccountKey } from "./support/accounts";
import { loginViaUi } from "./support/ui";

// Logs in once per account through the real login page and stores the browser
// session (access token in localStorage + refresh cookie) for the specs to reuse.
for (const key of Object.keys(accounts) as AccountKey[]) {
  setup(`log in as ${key}`, async ({ page }) => {
    await loginViaUi(page, accounts[key]);
    await page.context().storageState({ path: authFile(key) });
  });
}
