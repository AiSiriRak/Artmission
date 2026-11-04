import type { Page } from "@playwright/test";
import { createUser, removeUser, type NewUser } from "./support/api";
import { expect, test } from "./support/test";
import { loginViaUi } from "./support/ui";

const INVALID_LOGIN = "Cannot login because the email or password is incorrect.";

// Accounts without orders, created for the test that deletes (or almost deletes) them.
function newUser(email: string): NewUser {
  return {
    email,
    password: "Delete1234",
    username: "Delete",
    role: "customer",
    bankName: "kasikorn",
    accountHolderName: "Delete User",
    accountNumber: "1234567890",
  };
}

async function openProfile(page: Page, user: { email: string; password: string }) {
  await loginViaUi(page, user);
  await page.goto("/userprofile");
  await expect(page.getByRole("heading", { name: "User Bank Account Info" })).toBeVisible();
}

// The profile page and its confirmation popup both have a "Delete Account" button; the
// popup is rendered after the page, so its button is the last one.
const profileDelete = (page: Page) => page.getByRole("button", { name: "Delete Account" }).first();
const popupDelete = (page: Page) => page.getByRole("button", { name: "Delete Account" }).last();

async function expectCannotLogIn(page: Page, user: { email: string; password: string }) {
  await page.getByPlaceholder("Enter your email").fill(user.email);
  await page.getByPlaceholder("Enter your password").fill(user.password);
  await page.getByRole("button", { name: "Log in" }).click();
  await expect(page.getByText(INVALID_LOGIN)).toBeVisible();
}

async function deleteAndClose(page: Page, snap: (label: string) => Promise<void>) {
  await profileDelete(page).click();
  await popupDelete(page).click();
  await expect(page.getByText("Account Deleted")).toBeVisible();
  await expect(page.getByText("Your account has been successfully deleted.")).toBeVisible();
  await snap("popup");
  await page.getByRole("button", { name: "Close" }).click();
  await expect(page).toHaveURL(/\/login$/);
}

test("TC1-6-1 account without orders is deleted", async ({ page, request, snap }) => {
  const user = newUser("delete1@artmission.test");
  await createUser(request, user);
  try {
    await openProfile(page, user);
    await deleteAndClose(page, snap);
    await expectCannotLogIn(page, user);
  } finally {
    await removeUser(request, user);
  }
});

test("TC1-6-2 account whose only order is CANCEL is deleted", async ({ page, snap }) => {
  // Seeded artist with one cancelled order. Deleting it is permanent until the next reseed.
  const cleo = { email: "cleo@artmission.test", password: "Password123!" };
  await openProfile(page, cleo);
  await deleteAndClose(page, snap);
  await expectCannotLogIn(page, cleo);
});

test("TC1-6-3 account with ongoing orders cannot be deleted", async ({ page, snap }) => {
  await openProfile(page, { email: "jane@artmission.test", password: "Password123!" });
  await profileDelete(page).click();
  await popupDelete(page).click();

  await expect(page.getByText("Unable to Delete Account")).toBeVisible();
  await expect(
    page.getByText("Your account cannot be deleted at the moment, as there are still ongoing orders."),
  ).toBeVisible();
  await snap("popup");
  await page.getByRole("button", { name: "Close" }).click();

  await expect(page.getByText("Unable to Delete Account")).toHaveCount(0);
  await expect(page).toHaveURL(/\/userprofile$/);
  await expect(page.getByRole("heading", { name: "User Bank Account Info" })).toBeVisible();
});

test("TC1-6-4 cancelling the Delete Account popup keeps the account", async ({ page, request }) => {
  const user = newUser("delete2@artmission.test");
  await createUser(request, user);
  try {
    await openProfile(page, user);
    await profileDelete(page).click();
    await page.getByRole("button", { name: "Cancel" }).click();

    await expect(page.getByText("This action cannot be undone.", { exact: false })).toHaveCount(0);
    await expect(page).toHaveURL(/\/userprofile$/);
    await expect(page.getByRole("heading", { name: "User Bank Account Info" })).toBeVisible();
  } finally {
    await removeUser(request, user);
  }
});
