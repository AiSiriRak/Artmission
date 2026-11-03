import { test, expect } from "@playwright/test";
import { authFile } from "./support/accounts";
import { findOrderId } from "./support/api";

test.use({ storageState: authFile("artistBo") });

// Harness check, not a test case of the plan: proves the IN_PROCESS orders added to
// the seed are visible to the artist and look the way the order detail tests expect.
test.describe("seeded IN_PROCESS orders", () => {
  test("a plain order offers Postpone Order and Submit Draft", async ({ page, request }) => {
    const id = await findOrderId(request, "artistBo", "E2E draft - validation");
    await page.goto(`/orderhistory/${id}`);

    await expect(page.getByText("IN PROCESS", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: "Postpone Order" })).toBeEnabled();
    await expect(page.getByRole("heading", { name: "Submit Draft" })).toBeVisible();
  });

  test("an order with a pending extension request shows the banner", async ({ page, request }) => {
    const id = await findOrderId(request, "artistBo", "E2E postpone - request pending");
    await page.goto(`/orderhistory/${id}`);

    await expect(page.getByText("Postpone Request Pending")).toBeVisible();
    await expect(page.getByRole("button", { name: "Postpone Order" })).toBeDisabled();
  });
});
