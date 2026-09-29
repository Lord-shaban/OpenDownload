import { expect, test } from "@playwright/test";
test("analyze, select, queue and save the labeled fixture", async ({
  page,
}) => {
  await page.goto("/");
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/sample");
  await page.getByRole("button", { name: "Analyze link", exact: true }).click();
  await expect(
    page.getByRole("heading", {
      name: "A small film about the great outdoors",
    }),
  ).toBeVisible();
  await page.getByRole("radio", { name: /720p/ }).check();
  await page.getByRole("button", { name: "Queue download" }).click();
  await expect(page.getByText("Ready to save", { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("link", { name: "Save test file" }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toMatch(/\.txt$/);
  await page.getByRole("button", { name: /Delete A small/ }).click();
  await page
    .getByRole("button", { name: "Delete download", exact: true })
    .click();
  await expect(
    page.getByText("A little space for what you save."),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});
test("reject invalid URLs and expose recovery", async ({ page }) => {
  await page.goto("/");
  await page.getByLabel("Media link", { exact: true }).fill("file:///private");
  await page.getByRole("button", { name: "Analyze link" }).click();
  await expect(page.locator("#workspace-error")).toContainText(
    "Enter a complete public link",
  );
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/unsupported");
  await page.getByRole("button", { name: "Analyze link" }).click();
  await expect(page.locator("#workspace-error")).toContainText(
    "supported downloadable media",
  );
});
