import type { Page } from "@playwright/test";
import { authFile } from "./support/accounts";
import { findArtworkId } from "./support/api";
import { inputDate } from "./support/dates";
import { recordWrites } from "./support/network";
import { expect, test } from "./support/test";

test.use({ storageState: authFile("customerJane") });

// "Custom Portrait" has a minimum deadline of 5 days, so E (earliest valid deadline) is today + 6.
const E = 6;
const ORDER_NAME = "Portrait for my cat";
const DETAILS = "Black and white cat, watercolor style";
const MISSING_TEXT = "Please enter the order name and request details.";
const DEADLINE_TEXT = "Deadline must be at least 5 days from today.";

let orderPage: string;

test.beforeEach(async ({ page, request }) => {
  orderPage = `/artworks/${await findArtworkId(request, "customerJane", "Custom Portrait")}/order`;
  await page.goto(orderPage);
});

async function fillForm(page: Page, form: { name: string; details: string; deadline: string }) {
  await page.getByPlaceholder("Enter order name").fill(form.name);
  await page.getByPlaceholder("Enter request details").fill(form.details);
  await page.locator('input[type="date"]').fill(form.deadline);
}

// The form is rejected on the page: message shown, still on the order page, no create request sent.
async function expectRejected(page: Page, message: string) {
  const writes = recordWrites(page, "/orders");
  await page.getByRole("button", { name: "Create Order" }).click();

  await expect(page.getByText(message)).toBeVisible();
  await expect(page).toHaveURL(orderPage);
  expect(writes).toEqual([]);
}

test("TC6-1-1 valid order is created as PENDING", async ({ page }) => {
  await fillForm(page, { name: ORDER_NAME, details: DETAILS, deadline: inputDate(E) });
  await page.getByRole("button", { name: "Create Order" }).click();

  await expect(page).toHaveURL(/\/orderhistory\/[^/]+$/);
  await expect(page.getByText(ORDER_NAME, { exact: true })).toBeVisible();
  await expect(page.getByText("PENDING", { exact: true })).toBeVisible();
});

const missingText = [
  { id: "TC6-1-2", what: "empty Order Name", name: "", details: DETAILS },
  { id: "TC6-1-3", what: "Order Name of only spaces", name: "   ", details: DETAILS },
  { id: "TC6-1-4", what: "empty Request Details", name: ORDER_NAME, details: "" },
  { id: "TC6-1-5", what: "Request Details of only spaces", name: ORDER_NAME, details: "   " },
];

for (const { id, what, name, details } of missingText) {
  test(`${id} ${what} is rejected`, async ({ page }) => {
    await fillForm(page, { name, details, deadline: inputDate(E) });
    await expectRejected(page, MISSING_TEXT);
  });
}

const badDeadline = [
  { id: "TC6-1-6", what: "deadline one day before E", deadline: inputDate(E - 1) },
  { id: "TC6-1-7", what: "deadline in the past", deadline: inputDate(-1) },
  { id: "TC6-1-8", what: "cleared deadline", deadline: "" },
];

for (const { id, what, deadline } of badDeadline) {
  test(`${id} ${what} is rejected`, async ({ page }) => {
    await fillForm(page, { name: ORDER_NAME, details: DETAILS, deadline });
    await expectRejected(page, DEADLINE_TEXT);
  });
}
