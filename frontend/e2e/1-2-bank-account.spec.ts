import type { Locator } from "@playwright/test";
import { createUser, removeUser, type NewUser } from "./support/api";
import { recordWrites } from "./support/network";
import { expect, test } from "./support/test";
import { loginViaUi } from "./support/ui";

const KBANK = "ธนาคารกสิกรไทย (KBANK)";
const SCB = "ธนาคารไทยพาณิชย์ (SCB)";
const NUMBER_MISSING = "Please enter your bank account number.";
const NUMBER_INVALID = "Please enter a valid bank account number.";

// Each test edits an account of its own (the plan's john@gmail.com with the saved bank
// account of its precondition), so parallel tests cannot change each other's saved values.
let user: NewUser;
let bankCard: Locator;

test.beforeEach(async ({ page, request }, testInfo) => {
  const id = /^TC[\d-]+/.exec(testInfo.title)![0].toLowerCase();
  user = {
    email: `${id}@artmission.test`,
    password: "john12345",
    username: "John",
    role: "customer",
    bankName: "kasikorn",
    accountHolderName: "John Doe",
    accountNumber: "1234567890",
  };
  await createUser(request, user);
  await loginViaUi(page, user);
  await page.goto("/userprofile");

  bankCard = page.locator("div").filter({ has: page.getByRole("heading", { name: "User Bank Account Info" }) }).last();
  await bankCard.getByRole("button", { name: "Edit" }).click();
});

test.afterEach(async ({ request }) => {
  await removeUser(request, user);
});

async function fillBankForm(
  card: Locator,
  form: { bank: string; holder: string; number: string },
) {
  // The saved bank is KBANK; the dropdown shows it on its button.
  if (form.bank !== KBANK) {
    await card.getByRole("button", { name: KBANK }).click();
    await card.getByRole("button", { name: form.bank, exact: true }).click();
  }
  await card.getByPlaceholder("Enter account holder name").fill(form.holder);
  await card.getByPlaceholder("Enter account number").fill(form.number);
}

async function expectSavedValues(card: Locator, saved: { bank: string; holder: string; last4: string }) {
  await expect(card.getByText(saved.bank, { exact: true })).toBeVisible();
  await expect(card.getByText(saved.holder, { exact: true })).toBeVisible();
  await expect(card.getByText(saved.last4, { exact: true })).toBeVisible();
}

const ORIGINAL = { bank: KBANK, holder: "John Doe", last4: "••••7890" };

test("TC1-2-1 valid bank account is saved", async () => {
  await fillBankForm(bankCard, { bank: SCB, holder: "John Smith", number: "9876543210" });
  await bankCard.getByRole("button", { name: "Done" }).click();

  await expectSavedValues(bankCard, { bank: SCB, holder: "John Smith", last4: "••••3210" });
  await expect(bankCard.getByRole("button", { name: "Edit" })).toBeVisible();
});

const rejected: { id: string; what: string; holder: string; number: string; message: string }[] = [
  { id: "TC1-2-2", what: "cleared holder name", holder: "", number: "9876543210", message: "Please enter the account holder's name." },
  { id: "TC1-2-3", what: "holder name of only spaces", holder: "   ", number: "9876543210", message: "Please enter the account holder's name." },
  { id: "TC1-2-4", what: "empty account number", holder: "John Doe", number: "", message: NUMBER_MISSING },
  { id: "TC1-2-5", what: "account number with a letter", holder: "John Doe", number: "12345abc90", message: NUMBER_INVALID },
  { id: "TC1-2-6", what: "7-digit account number", holder: "John Doe", number: "1234567", message: NUMBER_INVALID },
  { id: "TC1-2-7", what: "17-digit account number", holder: "John Doe", number: "12345678901234567", message: NUMBER_INVALID },
];

for (const { id, what, holder, number, message } of rejected) {
  test(`${id} ${what} is rejected`, async ({ page }) => {
    const writes = recordWrites(page);
    await fillBankForm(bankCard, { bank: KBANK, holder, number });
    await bankCard.getByRole("button", { name: "Done" }).click();

    await expect(bankCard.getByText(message)).toBeVisible();
    await expect(bankCard.getByRole("button", { name: "Done" })).toBeVisible();
    expect(writes).toEqual([]);
  });
}

test("TC1-2-8 cancelling the edit keeps the saved values", async ({ page }) => {
  const writes = recordWrites(page);
  await fillBankForm(bankCard, { bank: SCB, holder: "Changed Name", number: "1111111111" });
  await bankCard.getByRole("button", { name: "Cancel" }).click();

  await expectSavedValues(bankCard, ORIGINAL);
  expect(writes).toEqual([]);
});
