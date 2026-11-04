import type { APIRequestContext, Page } from "@playwright/test";
import { removeUser } from "./support/api";
import { recordWrites } from "./support/network";
import { expect, test } from "./support/test";

interface Form {
  role?: "Customer" | "Artist";
  email: string;
  username: string;
  password: string;
  confirmPassword: string;
  bank?: string;
  holder: string;
  number: string;
  terms: boolean;
}

// Valid input set V of the test plan.
const V: Form = {
  role: "Customer",
  email: "john@gmail.com",
  username: "John",
  password: "john12345",
  confirmPassword: "john12345",
  bank: "ธนาคารกสิกรไทย (KBANK)",
  holder: "John Doe",
  number: "1234567890",
  terms: true,
};

const STEP_TITLE = { 1: "Choose your role", 2: "Account information", 3: "Bank account" } as const;

// Fills steps 1..lastStep and clicks Next (steps 1 and 2) or Create account (step 3), the
// way a user would; the flow of a case ends at the step where it is expected to stop.
async function fillRegistration(page: Page, form: Form, lastStep: 1 | 2 | 3) {
  await page.goto("/register");
  if (form.role) await page.getByRole("button", { name: new RegExp(`^${form.role}`) }).click();
  await page.getByRole("button", { name: /^Next/ }).click();
  if (lastStep === 1) return;

  await page.getByPlaceholder("Enter your email").fill(form.email);
  await page.getByPlaceholder("Enter your username").fill(form.username);
  await page.getByPlaceholder("Enter your password").fill(form.password);
  await page.getByPlaceholder("Confirm your password").fill(form.confirmPassword);
  await page.getByRole("button", { name: /^Next/ }).click();
  if (lastStep === 2) return;

  if (form.bank) await page.locator("select#bank").selectOption({ label: form.bank });
  await page.getByPlaceholder("Enter account holder name").fill(form.holder);
  await page.getByPlaceholder("Enter account number").fill(form.number);
  // The real checkbox is visually hidden; clicking the start of its label text toggles it.
  if (form.terms) await page.getByText("I agree to the").click({ position: { x: 2, y: 2 } });
  await page.getByRole("button", { name: "Create account" }).click();
}

async function expectAccountCreated(page: Page, snap: (label: string) => Promise<void>) {
  await expect(page.getByRole("heading", { name: "Account created!" })).toBeVisible();
  await expect(
    page.getByText("Your account has been created successfully. Please log in to continue."),
  ).toBeVisible();
  await snap("created");
  await page.getByRole("button", { name: "Log in" }).click();
  await expect(page).toHaveURL(/\/login$/);
}

async function createsAccount(
  page: Page,
  request: APIRequestContext,
  snap: (label: string) => Promise<void>,
  form: Form,
) {
  // A leftover of an earlier run would make the email "already registered".
  await removeUser(request, { email: form.email, password: form.password });
  await fillRegistration(page, form, 3);
  await expectAccountCreated(page, snap);
  await removeUser(request, { email: form.email, password: form.password });
}

test("TC1-1-1 valid customer registration", async ({ page, request, snap }) => {
  await createsAccount(page, request, snap, V);
});

test("TC1-1-2 valid artist registration, shortest values", async ({ page, request, snap }) => {
  await createsAccount(page, request, snap, {
    role: "Artist",
    email: "mei@gmail.com",
    username: "Mei",
    password: "mei12345",
    confirmPassword: "mei12345",
    bank: "ธนาคารไทยพาณิชย์ (SCB)",
    holder: "Mei Tan",
    number: "12345678",
    terms: true,
  });
});

test("TC1-1-3 valid customer registration, longest values", async ({ page, request, snap }) => {
  await createsAccount(page, request, snap, {
    role: "Customer",
    email: "max@gmail.com",
    username: "maximumlengthuser123",
    password: "max1234567890123",
    confirmPassword: "max1234567890123",
    bank: "United Overseas Bank (UOB)",
    holder: "Max Maximum",
    number: "1234567890123456",
    terms: true,
  });
});

// Input rejected by the form: it stays on `step`, shows `message`, and sends nothing.
const rejected: { id: string; what: string; change: Partial<Form>; step: 1 | 2 | 3; message: string }[] = [
  { id: "TC1-1-4", what: "no role selected", change: { role: undefined }, step: 1, message: "Please select your role." },
  { id: "TC1-1-5", what: "empty email", change: { email: "" }, step: 2, message: "Please enter your email address." },
  { id: "TC1-1-6", what: "malformed email", change: { email: "john.gmail.com" }, step: 2, message: "Please enter a valid email address." },
  { id: "TC1-1-8", what: "empty username", change: { username: "" }, step: 2, message: "Please enter your username." },
  { id: "TC1-1-9", what: "username of only spaces", change: { username: "   " }, step: 2, message: "Please enter your username." },
  { id: "TC1-1-10", what: "2-character username", change: { username: "Jo" }, step: 2, message: "Username must be at least 3 characters." },
  { id: "TC1-1-11", what: "21-character username", change: { username: "maximumlengthuser1234" }, step: 2, message: "Username must be 20 characters or less." },
  { id: "TC1-1-12", what: "empty password", change: { password: "", confirmPassword: "" }, step: 2, message: "Please enter a password." },
  { id: "TC1-1-13", what: "7-character password", change: { password: "john123", confirmPassword: "john123" }, step: 2, message: "Password must be at least 8 characters." },
  { id: "TC1-1-14", what: "17-character password", change: { password: "john1234567890123", confirmPassword: "john1234567890123" }, step: 2, message: "Password must be 16 characters or less." },
  { id: "TC1-1-15", what: "different confirm password", change: { confirmPassword: "john54321" }, step: 2, message: "Passwords do not match." },
  { id: "TC1-1-16", what: "no bank selected", change: { bank: undefined }, step: 3, message: "Please select your bank." },
  { id: "TC1-1-17", what: "empty account holder name", change: { holder: "" }, step: 3, message: "Please enter the account holder's name." },
  { id: "TC1-1-18", what: "account holder name of only spaces", change: { holder: "   " }, step: 3, message: "Please enter the account holder's name." },
  { id: "TC1-1-19", what: "empty account number", change: { number: "" }, step: 3, message: "Please enter your bank account number." },
  { id: "TC1-1-20", what: "account number with a letter", change: { number: "12345abc90" }, step: 3, message: "Please enter a valid bank account number." },
  { id: "TC1-1-21", what: "7-digit account number", change: { number: "1234567" }, step: 3, message: "Please enter a valid bank account number." },
  { id: "TC1-1-22", what: "17-digit account number", change: { number: "12345678901234567" }, step: 3, message: "Please enter a valid bank account number." },
  { id: "TC1-1-23", what: "terms not accepted", change: { terms: false }, step: 3, message: "Please agree to the Terms of Service and Privacy Policy to continue." },
];

for (const { id, what, change, step, message } of rejected) {
  test(`${id} ${what} is rejected`, async ({ page }) => {
    const writes = recordWrites(page);
    await fillRegistration(page, { ...V, ...change }, step);

    await expect(page.getByRole("heading", { name: STEP_TITLE[step] })).toBeVisible();
    await expect(page.getByText(message)).toBeVisible();
    expect(writes).toEqual([]);
  });
}

test("TC1-1-7 already registered email is rejected", async ({ page }) => {
  await fillRegistration(page, { ...V, email: "jane@artmission.test" }, 3);

  // The server refuses it, so the form returns to step 2.
  await expect(page.getByRole("heading", { name: STEP_TITLE[2] })).toBeVisible();
  await expect(page.getByText("This email is already registered.")).toBeVisible();
  await expect(page.getByRole("heading", { name: "Account created!" })).toHaveCount(0);
});
