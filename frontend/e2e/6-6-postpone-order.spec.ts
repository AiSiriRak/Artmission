import type { APIRequestContext, Locator, Page } from "@playwright/test";
import { authFile } from "./support/accounts";
import { findOrderId } from "./support/api";
import { displayDate, inputDate } from "./support/dates";
import { recordWrites } from "./support/network";
import { expect, test } from "./support/test";

test.use({ storageState: authFile("artistBo") });

// The seeded orders are due in 14 days (D), so the new deadline D + 7 is today + 21.
const D = 14;
const REASON = "I need more time for details";
const EXPLANATION = "Need extra time to refine the shading.";
const ALERT = "กรุณาเลือกวันหลัง Deadline เดิม พร้อมเหตุผลและคำอธิบาย";

async function openOrder(
  page: Page,
  request: APIRequestContext,
  orderName: string,
) {
  await page.goto(
    `/orderhistory/${await findOrderId(request, "artistBo", orderName)}`,
  );
}

async function openPopup(page: Page): Promise<Locator> {
  await page.getByRole("button", { name: "Postpone Order" }).click();
  const popup = page.getByRole("dialog", {
    name: "Request to Postpone Order?",
  });
  await expect(popup).toBeVisible();
  return popup;
}

async function fillPopup(
  popup: Locator,
  form: { deadline?: string; reason?: string; explanation?: string },
) {
  if (form.deadline)
    await popup.locator('input[type="date"]').fill(form.deadline);
  if (form.reason) {
    await popup
      .getByRole("button", { name: "Select reason for postpone" })
      .click();
    await popup.getByRole("button", { name: form.reason, exact: true }).click();
  }
  if (form.explanation) await popup.locator("textarea").fill(form.explanation);
}

async function send(popup: Locator) {
  await popup.getByRole("button", { name: "Send Request" }).click();
}

async function expectNotSent(page: Page, popup: Locator, writes: string[]) {
  await expect(popup).toBeVisible();
  await expect(page.getByText("Postpone Request Pending")).toHaveCount(0);
  expect(writes).toEqual([]);
}

// The browser blocks the submit and shows its own validation message at the field.
async function expectBrowserValidation(field: Locator) {
  expect(
    await field.evaluate(
      (element) => (element as HTMLInputElement).validationMessage,
    ),
  ).not.toBe("");
}

test("TC6-6-1 valid request is sent", async ({ page, request }) => {
  await openOrder(page, request, "E2E postpone - send request");
  const popup = await openPopup(page);
  await fillPopup(popup, {
    deadline: inputDate(D + 7),
    reason: REASON,
    explanation: EXPLANATION,
  });
  await send(popup);

  await expect(popup).toBeHidden();
  const banner = page
    .getByRole("status")
    .filter({ hasText: "Postpone Request Pending" });
  await expect(banner).toContainText(
    `You have requested to extend the deadline to ${displayDate(D + 7)}. Waiting for the customer's response (Auto declines in 48 hours).`,
  );
  await expect(
    page.getByRole("button", { name: "Postpone Order" }),
  ).toBeDisabled();
});

const badDeadline = [
  {
    id: "TC6-6-2",
    what: "same date as the current deadline",
    deadline: inputDate(D),
  },
  {
    id: "TC6-6-3",
    what: "date before the current deadline",
    deadline: inputDate(D - 1),
  },
  { id: "TC6-6-4", what: "empty date", deadline: undefined },
];

for (const { id, what, deadline } of badDeadline) {
  test(`${id} ${what} is not sent`, async ({ page, request }) => {
    await openOrder(page, request, "E2E postpone - validation");
    const popup = await openPopup(page);
    const writes = recordWrites(page);
    await fillPopup(popup, {
      deadline,
      reason: REASON,
      explanation: EXPLANATION,
    });
    await send(popup);

    await expectBrowserValidation(popup.locator('input[type="date"]'));
    await expectNotSent(page, popup, writes);
  });
}

test("TC6-6-5 missing reason is not sent", async ({ page, request }) => {
  await openOrder(page, request, "E2E postpone - validation");
  const popup = await openPopup(page);
  const writes = recordWrites(page);
  await fillPopup(popup, {
    deadline: inputDate(D + 7),
    explanation: EXPLANATION,
  });
  await send(popup);

  await expect(popup.getByRole("alert")).toHaveText(ALERT);
  await expectNotSent(page, popup, writes);
});

test("TC6-6-6 empty explanation is not sent", async ({ page, request }) => {
  await openOrder(page, request, "E2E postpone - validation");
  const popup = await openPopup(page);
  const writes = recordWrites(page);
  await fillPopup(popup, { deadline: inputDate(D + 7), reason: REASON });
  await send(popup);

  await expectBrowserValidation(popup.locator("textarea"));
  await expectNotSent(page, popup, writes);
});

test("TC6-6-7 explanation of only spaces is not sent", async ({
  page,
  request,
}) => {
  await openOrder(page, request, "E2E postpone - validation");
  const popup = await openPopup(page);
  const writes = recordWrites(page);
  await fillPopup(popup, {
    deadline: inputDate(D + 7),
    reason: REASON,
    explanation: "   ",
  });
  await send(popup);

  await expect(popup.getByRole("alert")).toHaveText(ALERT);
  await expectNotSent(page, popup, writes);
});

test("TC6-6-8 explanation is limited to 300 characters", async ({
  page,
  request,
}) => {
  await openOrder(page, request, "E2E postpone - validation");
  const popup = await openPopup(page);
  await fillPopup(popup, { deadline: inputDate(D + 7), reason: REASON });
  // fill() would bypass maxLength; typing is what a user does.
  await popup.locator("textarea").pressSequentially("A".repeat(301));

  await expect(popup.getByText("300/300")).toBeVisible();
  await expect(popup.locator("textarea")).toHaveValue("A".repeat(300));
});

test("TC6-6-9 NOT_PAID order offers no Postpone Order", async ({
  page,
  request,
}) => {
  await openOrder(page, request, "E2E not paid order");

  await expect(page.getByText("NOT PAID", { exact: true })).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Postpone Order" }),
  ).toHaveCount(0);
});

test("TC6-6-10 pending request blocks a second one", async ({
  page,
  request,
}) => {
  await openOrder(page, request, "E2E postpone - request pending");

  await expect(page.getByText("Postpone Request Pending")).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Postpone Order" }),
  ).toBeDisabled();
});
