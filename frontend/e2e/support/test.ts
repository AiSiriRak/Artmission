import { test as base, type Page } from "@playwright/test";

export { expect } from "@playwright/test";

// A screenshot of the whole page. The viewport is made as tall as the page instead of using
// fullPage, so sticky headers and modal overlays are drawn where a user sees them.
async function capture(page: Page, name: string) {
  const viewport = page.viewportSize();
  if (!viewport || page.isClosed()) return;
  await page.evaluate(() => window.scrollTo(0, 0));
  const height = await page.evaluate(() => document.documentElement.scrollHeight);
  await page.setViewportSize({ width: viewport.width, height: Math.max(viewport.height, height) });
  await page.screenshot({ path: `e2e-evidence/${name}.png` });
  await page.setViewportSize(viewport);
}

const caseId = (title: string) => /^TC[\d-]+/.exec(title)?.[0];

// Every spec imports `test` from here instead of "@playwright/test". After each test it
// saves e2e-evidence/<test case ID>.png, the final state of the page, so the result of every
// test case can be pasted into the test plan. A test whose expected output appears before
// its last step also calls `snap("label")` at that moment, which saves
// e2e-evidence/<test case ID>-<label>.png.
export const test = base.extend<{ evidence: void; snap: (label: string) => Promise<void> }>({
  snap: async ({ page }, use, testInfo) => {
    await use((label) => capture(page, `${caseId(testInfo.title)}-${label}`));
  },
  evidence: [
    async ({ page }, use, testInfo) => {
      await use();
      const id = caseId(testInfo.title);
      if (id) await capture(page, id);
    },
    { auto: true },
  ],
});
