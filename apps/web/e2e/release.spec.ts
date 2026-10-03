import { expect, test } from "@playwright/test";

for (const locale of ["en", "ar"] as const) {
  test(`v0.1 YouTube boundary and recovery in ${locale}`, async ({ page }) => {
    await page.goto("/");
    await expect(
      page.getByText("Fixture mode.", { exact: false }),
    ).toBeVisible();
    if (locale === "ar")
      await page.getByLabel("Language", { exact: true }).selectOption("ar");
    await page.locator("details.supported-sites > summary").click();
    const list = page.getByRole("list", {
      name: locale === "ar" ? "المواقع المدعومة" : "Supported sites",
    });
    await expect(list).not.toContainText("YouTube");
    await expect(list).toContainText("TikTok");
    for (const source of ["LinkedIn", "Pinterest", "Threads"])
      await expect(list).toContainText(source);
    if (test.info().project.name === "mobile") {
      await page.screenshot({
        path: `../../.data/release-0.1-sites-${locale}.png`,
        fullPage: true,
      });
    }
    const input = page.getByLabel(
      locale === "ar" ? "رابط الوسائط" : "Media link",
      { exact: true },
    );
    await input.fill("https://youtu.be/2cUkUbB3Gu4");
    await input.press("Enter");
    await expect(page.locator("#workspace-error")).toContainText(
      locale === "ar"
        ? "YouTube غير مدعوم في الإصدار 0.1"
        : "YouTube is not supported in v0.1",
    );
    await expect(page.getByRole("radio")).toHaveCount(0);
    await input.fill("https://example.com/sample");
    await input.press("Enter");
    await expect(page.getByRole("radio", { name: /1080p/ })).toBeChecked();
    await expect(page.locator("#workspace-error")).toHaveCount(0);
  });
}
