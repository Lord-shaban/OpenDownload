import { expect, test } from "@playwright/test";

test("persist theme and guide Enter → format → download with useful focus", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Switch to dark theme" }).click();
  await page.reload();
  await expect(page.locator("html")).toHaveClass(/dark/);
  await expect(
    page.getByRole("button", { name: "Switch to light theme" }),
  ).toBeVisible();
  const serverHTML = await (await page.request.get("/")).text();
  expect(serverHTML).toMatch(/<html[^>]+class="[^"]*dark/);
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/sample");
  await page.getByLabel("Media link", { exact: true }).press("Enter");
  await expect(
    page.getByLabel("Analysis results", { exact: true }),
  ).toBeFocused();
  await expect(
    page
      .getByRole("list", { name: "Download steps" })
      .locator('[aria-current="step"]'),
  ).toContainText("Choose a format");
  await page.getByRole("button", { name: "Download", exact: true }).click();
  await expect(
    page
      .getByRole("navigation", { name: "Main navigation" })
      .getByRole("button", { name: "Downloads", exact: true }),
  ).toHaveAttribute("aria-current", "page");
  await expect(page.locator("#main-content")).toBeFocused();
  await expect(
    page
      .getByRole("list", { name: "Download steps" })
      .locator('[aria-current="step"]'),
  ).toContainText("Process & save");
  await expect(
    page.getByText("Your file is ready", { exact: true }),
  ).toBeVisible({
    timeout: 15_000,
  });
  await page
    .getByRole("navigation")
    .getByRole("button", { name: "New download", exact: true })
    .click();
  await expect(page.getByLabel("Media link", { exact: true })).toBeFocused();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});

test("clipboard button and native paste analyze without an extra submit", async ({
  page,
  context,
}) => {
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.goto("/");
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  const supportedSites = page.locator("details.supported-sites > summary");
  await supportedSites.focus();
  await supportedSites.press("Enter");
  await expect(
    page.getByRole("list", { name: "Supported sites" }),
  ).toBeVisible();
  const sourceList = page.getByRole("list", { name: "Supported sites" });
  for (const source of ["YouTube", "Instagram", "TikTok", "SoundCloud"]) {
    await expect(sourceList).toContainText(source);
  }
  await expect(
    page.getByText(/Photo galleries have limited support/),
  ).toBeVisible();
  await supportedSites.press("Enter");
  await page.evaluate(async () => {
    await navigator.clipboard.writeText("https://example.com/sample");
  });
  await page
    .getByRole("button", { name: "Paste link from clipboard", exact: true })
    .click();
  await expect(
    page.getByLabel("Analysis results", { exact: true }),
  ).toBeFocused();
  await expect(page.getByRole("radio", { name: /1080p/ })).toBeChecked();
  // Choosing a format and downloading remain deliberate actions.
  await expect(page.getByRole("list", { name: "Download jobs" })).toHaveCount(
    0,
  );
  await page.getByRole("button", { name: "Clear link", exact: true }).click();
  await page.evaluate(async () => {
    await navigator.clipboard.writeText("https://example.com/unsupported");
  });
  await page.getByLabel("Media link", { exact: true }).focus();
  await page.keyboard.press("Control+V");
  await expect(page.locator("#workspace-error")).toContainText(
    "Couldn't reach public media",
  );
  await expect(
    page.getByLabel("Analysis results", { exact: true }),
  ).toHaveCount(0);
});

test("clipboard denial explains keyboard paste and returns focus", async ({
  page,
}) => {
  await page.addInitScript(() => {
    Object.defineProperty(navigator, "clipboard", {
      value: {
        readText: async () => {
          throw new DOMException("Denied", "NotAllowedError");
        },
      },
    });
  });
  await page.goto("/");
  await expect(page.getByText("Fixture mode.", { exact: false })).toBeVisible();
  await page
    .getByRole("button", { name: "Paste link from clipboard", exact: true })
    .click();
  await expect(page.locator("#workspace-error")).toContainText("Ctrl+V");
  await expect(page.getByLabel("Media link", { exact: true })).toBeFocused();
  await page
    .getByLabel("Media link", { exact: true })
    .fill("https://example.com/sample");
  await page.getByLabel("Media link", { exact: true }).press("Enter");
  await expect(
    page.getByLabel("Analysis results", { exact: true }),
  ).toBeFocused();
});
