import type { Page } from "@playwright/test";

// Logs in through the login page, the way a user does.
export async function loginViaUi(page: Page, user: { email: string; password: string }) {
  await page.goto("/login");
  await page.getByPlaceholder("Enter your email").fill(user.email);
  await page.getByPlaceholder("Enter your password").fill(user.password);
  await page.getByRole("button", { name: "Log in" }).click();
  // Where a user lands depends on the role, so only wait until the login page is left.
  await page.waitForURL((url) => url.pathname !== "/login");
}
