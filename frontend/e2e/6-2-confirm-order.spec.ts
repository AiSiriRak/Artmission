import type { Page } from "@playwright/test";
import { authFile } from "./support/accounts";
import { findOrderId, placeOrder } from "./support/api";
import { expect, test } from "./support/test";

const accept = (page: Page) => page.getByRole("button", { name: "Accept", exact: true });
const reject = (page: Page) => page.getByRole("button", { name: "Reject", exact: true });

test.describe("PENDING order, owned by ann", () => {
  test.use({ storageState: authFile("artistAnn") });

  // Each test places its own PENDING order through the real API, so they never share state.
  test.beforeEach(async ({ page, request }, testInfo) => {
    const order = await placeOrder(request, { name: testInfo.title });
    await page.goto(`/orderhistory/${order.id}`);
    await expect(page.getByText("PENDING", { exact: true })).toBeVisible();
  });

  test("TC6-2-1 artist accepts a PENDING order", async ({ page }) => {
    await accept(page).click();
    const popup = page.getByRole("dialog", { name: "Accept this Order?" });
    await popup.getByRole("button", { name: "Accept", exact: true }).click();

    await expect(popup).toBeHidden();
    await expect(page.getByText("NOT PAID", { exact: true })).toBeVisible();
    await expect(accept(page)).toHaveCount(0);
    await expect(reject(page)).toHaveCount(0);
  });

  test("TC6-2-2 artist rejects a PENDING order", async ({ page }) => {
    await reject(page).click();
    const popup = page.getByRole("dialog", { name: "Reject this Order?" });
    await popup.getByRole("button", { name: "Reject", exact: true }).click();

    await expect(popup).toBeHidden();
    await expect(page.getByText("CANCEL", { exact: true })).toBeVisible();
    await expect(accept(page)).toHaveCount(0);
    await expect(reject(page)).toHaveCount(0);
  });

  test("TC6-2-3 artist cancels the Accept popup", async ({ page }) => {
    await accept(page).click();
    const popup = page.getByRole("dialog", { name: "Accept this Order?" });
    await popup.getByRole("button", { name: "Cancel" }).click();

    await expect(popup).toBeHidden();
    await expect(page.getByText("PENDING", { exact: true })).toBeVisible();
    await expect(accept(page)).toBeVisible();
    await expect(reject(page)).toBeVisible();
  });

  test("TC6-2-4 artist cancels the Reject popup", async ({ page }) => {
    await reject(page).click();
    const popup = page.getByRole("dialog", { name: "Reject this Order?" });
    await popup.getByRole("button", { name: "Cancel" }).click();

    await expect(popup).toBeHidden();
    await expect(page.getByText("PENDING", { exact: true })).toBeVisible();
    await expect(accept(page)).toBeVisible();
    await expect(reject(page)).toBeVisible();
  });
});

// Seeded orders of bo that are not PENDING; they are only opened, never changed.
test.describe("order that is not PENDING", () => {
  test.use({ storageState: authFile("artistBo") });

  const cases = [
    { id: "TC6-2-5", orderName: "E2E not paid order", badge: "NOT PAID" },
    { id: "TC6-2-6", orderName: "E2E postpone - validation", badge: "IN PROCESS" },
    { id: "TC6-2-7", orderName: "E2E completed order", badge: "SUCCESS" },
    { id: "TC6-2-8", orderName: "E2E cancelled order", badge: "CANCEL" },
  ];

  for (const { id, orderName, badge } of cases) {
    test(`${id} ${badge} order offers no Accept or Reject`, async ({ page, request }) => {
      await page.goto(`/orderhistory/${await findOrderId(request, "artistBo", orderName)}`);

      await expect(page.getByText(badge, { exact: true })).toBeVisible();
      await expect(accept(page)).toHaveCount(0);
      await expect(reject(page)).toHaveCount(0);
    });
  }
});
