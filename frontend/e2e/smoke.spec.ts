import { test, expect } from "@playwright/test";
import { authFile } from "./support/accounts";
import { placeOrder } from "./support/api";

test.use({ storageState: authFile("artistAnn") });

// Smoke test of the whole harness: seeded login, a fresh PENDING order created
// through the real API, and one full UI flow.
test("TC6-2-1 artist accepts a PENDING order", async ({ page, request }) => {
  const order = await placeOrder(request, { name: `TC6-2-1 ${Date.now()}` });

  await page.goto(`/orderhistory/${order.id}`);
  await expect(page.getByText("PENDING", { exact: true })).toBeVisible();

  await page.getByRole("button", { name: "Accept", exact: true }).click();
  const popup = page.getByRole("dialog", { name: "Accept this Order?" });
  await popup.getByRole("button", { name: "Accept", exact: true }).click();

  await expect(popup).toBeHidden();
  await expect(page.getByText("NOT PAID", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Accept", exact: true })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Reject", exact: true })).toHaveCount(0);
});
