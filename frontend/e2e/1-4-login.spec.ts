import type { Page } from "@playwright/test";
import { recordWrites } from "./support/network";
import { expect, test } from "./support/test";

const INVALID_LOGIN = "Cannot login because the email or password is incorrect.";
const PASSWORD = "Password123!";

async function fillLogin(page: Page, email: string, password: string) {
  await page.goto("/login");
  await page.getByPlaceholder("Enter your email").fill(email);
  await page.getByPlaceholder("Enter your password").fill(password);
}

const logIn = (page: Page) => page.getByRole("button", { name: "Log in" });

const loggedIn = [
  { id: "TC1-4-1", who: "customer", email: "jane@artmission.test", nav: ["Home", "Order", "Notification"] },
  { id: "TC1-4-2", who: "artist", email: "ann@artmission.test", nav: ["Order", "Artist Profile", "Notification"] },
];

for (const { id, who, email, nav } of loggedIn) {
  test(`${id} ${who} logs in with the correct password`, async ({ page }) => {
    await fillLogin(page, email, PASSWORD);
    await logIn(page).click();

    // A customer stays on the home page; an artist is sent on to the order history.
    if (who === "customer") {
      await expect(page).toHaveURL("/");
      await expect(page.getByRole("heading", { name: "Browse artists" })).toBeVisible();
    }
    for (const item of nav) {
      await expect(page.getByRole("navigation").getByRole("link", { name: item })).toBeVisible();
    }
  });
}

const refused = [
  { id: "TC1-4-3", what: "wrong password", email: "jane@artmission.test", password: "Password123" },
  { id: "TC1-4-4", what: "unregistered email", email: "nobody@artmission.test", password: PASSWORD },
];

for (const { id, what, email, password } of refused) {
  test(`${id} ${what} is refused`, async ({ page }) => {
    await fillLogin(page, email, password);
    await logIn(page).click();

    await expect(page.getByText(INVALID_LOGIN)).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);
  });
}

test("TC1-4-5 malformed email is blocked by the browser", async ({ page }) => {
  const writes = recordWrites(page);
  await fillLogin(page, "jane.artmission.test", PASSWORD);
  await logIn(page).click();

  const email = page.getByPlaceholder("Enter your email");
  expect(await email.evaluate((element) => (element as HTMLInputElement).validationMessage)).not.toBe("");
  await expect(page).toHaveURL(/\/login$/);
  expect(writes).toEqual([]);
});

const disabled = [
  { id: "TC1-4-6", what: "empty email", email: "", password: PASSWORD },
  { id: "TC1-4-7", what: "empty password", email: "jane@artmission.test", password: "" },
];

for (const { id, what, email, password } of disabled) {
  test(`${id} ${what} disables Log in`, async ({ page }) => {
    await fillLogin(page, email, password);

    await expect(logIn(page)).toBeDisabled();
    await expect(page).toHaveURL(/\/login$/);
  });
}
