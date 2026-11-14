import { readFileSync } from "node:fs";
import type { APIRequestContext, Page } from "@playwright/test";
import { authFile } from "./support/accounts";
import { findOrderId } from "./support/api";
import { expect, test } from "./support/test";

test.use({ storageState: authFile("artistBo") });

const MiB = 1024 * 1024;
const fixture = (name: string) => readFileSync(`e2e/fixtures/${name}`);

async function openOrder(page: Page, request: APIRequestContext, orderName: string) {
  await page.goto(`/orderhistory/${await findOrderId(request, "artistBo", orderName)}`);
}

const chooseFile = (page: Page, file: { name: string; mimeType: string; buffer: Buffer }) =>
  page.getByLabel("Choose draft image").setInputFiles(file);

const submitButton = (page: Page) => page.getByRole("button", { name: "Submit", exact: true });
const confirmPopup = (page: Page) => page.getByRole("dialog", { name: "Submit Draft to Customer?" });
const submitHeading = (page: Page) => page.getByRole("heading", { name: "Submit Draft" });

async function expectDraftSubmitted(page: Page) {
  await expect(confirmPopup(page)).toBeHidden();
  await expect(page.getByText("#1 Draft")).toBeVisible();
  await expect(page.getByText("Waiting for Customer Review")).toBeVisible();
  await expect(submitHeading(page)).toHaveCount(0);
}

// The server refuses the file: popup stays open with the error, no draft card appears.
async function expectServerRejected(page: Page, message: string) {
  await submitButton(page).click();
  const popup = confirmPopup(page);
  await popup.getByRole("button", { name: "Submit", exact: true }).click();

  await expect(popup.getByRole("alert")).toContainText(message);
  await expect(page.getByText("#1 Draft")).toHaveCount(0);
}

const validTypes = [
  { id: "TC6-7-1", order: "E2E draft - JPEG", file: "draft.jpg", mimeType: "image/jpeg" },
  { id: "TC6-7-2", order: "E2E draft - PNG", file: "draft.png", mimeType: "image/png" },
  { id: "TC6-7-3", order: "E2E draft - WebP", file: "draft.webp", mimeType: "image/webp" },
];

for (const { id, order, file, mimeType } of validTypes) {
  test(`${id} ${file} is submitted`, async ({ page, request }) => {
    await openOrder(page, request, order);
    await chooseFile(page, { name: file, mimeType, buffer: fixture(file) });
    await submitButton(page).click();
    await confirmPopup(page).getByRole("button", { name: "Submit", exact: true }).click();

    await expectDraftSubmitted(page);
  });
}

test("TC6-7-4 GIF file is refused", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - validation");
  await chooseFile(page, { name: "draft.gif", mimeType: "image/gif", buffer: Buffer.from("GIF89a") });

  await expect(page.getByRole("alert").filter({ hasText: "เลือกไฟล์ JPG, PNG หรือ WebP เท่านั้น" })).toBeVisible();
  await expect(page.getByAltText("Draft preview")).toHaveCount(0);
  await expect(submitButton(page)).toHaveCount(0);
});

test("TC6-7-5 no file chosen offers no Submit", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - validation");

  await expect(page.getByText("Add your artwork")).toBeVisible();
  await expect(submitButton(page)).toHaveCount(0);
});

test("TC6-7-6 text file renamed to .png is rejected by the server", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - validation");
  await chooseFile(page, { name: "fake.png", mimeType: "image/png", buffer: Buffer.from("this is not an image") });

  await expectServerRejected(page, "deliverable image must be a JPEG, PNG, or WebP image");
});

test("TC6-7-7 image over 20 MiB is rejected by the server", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - validation");
  // A valid PNG followed by padding: the server compares the size before it decodes.
  const big = Buffer.concat([fixture("draft.png"), Buffer.alloc(21 * MiB - fixture("draft.png").length)]);
  await chooseFile(page, { name: "big.png", mimeType: "image/png", buffer: big });

  await expectServerRejected(page, "deliverable image exceeds the 20 MiB limit");
});

test("TC6-7-8 cancelling the Submit popup keeps the file", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - validation");
  await chooseFile(page, { name: "draft.jpg", mimeType: "image/jpeg", buffer: fixture("draft.jpg") });
  await submitButton(page).click();
  await confirmPopup(page).getByRole("button", { name: "Cancel" }).click();

  await expect(confirmPopup(page)).toBeHidden();
  await expect(page.getByAltText("Draft preview")).toBeVisible();
  await expect(submitButton(page)).toBeVisible();
  await expect(page.getByText("#1 Draft")).toHaveCount(0);
});

test("TC6-7-9 removing the chosen file restores the upload area", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - validation");
  await chooseFile(page, { name: "draft.jpg", mimeType: "image/jpeg", buffer: fixture("draft.jpg") });
  await page.getByRole("button", { name: "Remove draft image" }).click();

  await expect(page.getByAltText("Draft preview")).toHaveCount(0);
  await expect(page.getByText("Add your artwork")).toBeVisible();
  await expect(submitButton(page)).toHaveCount(0);
});

test("TC6-7-10 NOT_PAID order has no Submit Draft section", async ({ page, request }) => {
  await openOrder(page, request, "E2E not paid order");

  await expect(page.getByText("NOT PAID", { exact: true })).toBeVisible();
  await expect(submitHeading(page)).toHaveCount(0);
});

test("TC6-7-11 draft awaiting review has no Submit Draft section", async ({ page, request }) => {
  await openOrder(page, request, "E2E draft - awaiting review");

  await expect(page.getByText("Waiting for Customer Review")).toBeVisible();
  await expect(submitHeading(page)).toHaveCount(0);
});
