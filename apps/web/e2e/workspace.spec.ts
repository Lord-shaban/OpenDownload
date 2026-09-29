import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";
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
  await page.getByText("720p", { exact: true }).click();
  await expect(page.getByRole("radio", { name: /720p/ })).toBeChecked();
  await page.getByRole("button", { name: "Queue download" }).click();
  await expect(page.getByText("Ready to save", { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("link", { name: "Save test file" }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toMatch(/\.txt$/);
  const saved = await download.path();
  expect(await readFile(saved!, "utf8")).toContain(
    "It is not extracted media.",
  );
  await page.getByRole("button", { name: /Delete .*A small/ }).click();
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
test("cancel active work and retry as a new job", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/sample");
  await page.getByRole("button", { name: "Analyze link", exact: true }).click();
  await page.getByRole("button", { name: "Queue download" }).click();
  await page.getByRole("button", { name: /^Cancel .*A small/ }).click();
  await expect(page.getByText("Canceled", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: /^Retry .*A small/ }).click();
  await expect(page.getByText("Ready to save", { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await expect(
    page.getByRole("list", { name: "Download jobs" }).getByRole("listitem"),
  ).toHaveCount(2);
  await expect(page.getByText("Canceled", { exact: true })).toBeVisible();
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
test("show processing failure and retry without inventing success", async ({
  page,
}) => {
  await page.goto("/");
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/fail");
  await page.getByRole("button", { name: "Analyze link", exact: true }).click();
  await page.getByRole("button", { name: "Queue download" }).click();
  await expect(page.getByText("Couldn’t finish", { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await page.getByRole("button", { name: /^Retry .*A small/ }).click();
  await expect(page.getByText("Couldn’t finish", { exact: true })).toHaveCount(
    2,
    { timeout: 15_000 },
  );
  await expect(page.getByRole("link", { name: "Save test file" })).toHaveCount(
    0,
  );
});
test("reflow, reduced motion and keyboard dialog dismissal", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.setViewportSize({ width: 375, height: 812 });
  await page.goto("/");
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/sample");
  await page.getByRole("button", { name: "Analyze link", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Queue download" }),
  ).toBeVisible();
  expect(
    await page
      .locator(".panel-enter")
      .evaluate((el) => getComputedStyle(el).animationName),
  ).toBe("none");
  for (const size of [
    { width: 375, height: 812 },
    { width: 812, height: 375 },
    { width: 768, height: 1024 },
  ]) {
    await page.setViewportSize(size);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
  }
  await page.getByRole("button", { name: "Help", exact: true }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Help", exact: true }),
  ).toBeFocused();
});
