import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";

test("switch locale without losing input, persist SSR direction and return to English", async ({
  page,
}) => {
  await page.goto("/");
  const source = "https://example.com/فيلم?format=mp4";
  await page.getByLabel("Media link", { exact: true }).fill(source);
  await page.getByLabel("Language", { exact: true }).selectOption("ar");
  await expect(page.locator("html")).toHaveAttribute("dir", "rtl");
  await expect(
    page
      .getByRole("navigation")
      .getByRole("button", { name: "تنزيل جديد", exact: true }),
  ).toBeVisible();
  await expect(page.getByLabel("رابط الوسائط", { exact: true })).toHaveValue(
    source,
  );
  await expect(
    page.getByLabel("رابط الوسائط", { exact: true }),
  ).toHaveAttribute("dir", "ltr");
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("lang", "ar");
  const serverHTML = await (await page.request.get("/")).text();
  expect(serverHTML).toMatch(/<html[^>]+lang="ar"[^>]+dir="rtl"/);
  expect(serverHTML).toContain("وسائطك، بطريقتك.");
  await page.getByLabel("اللغة", { exact: true }).selectOption("en");
  await expect(page.locator("html")).toHaveAttribute("dir", "ltr");
  await expect(
    page
      .getByRole("navigation")
      .getByRole("button", { name: "New download", exact: true }),
  ).toBeVisible();
});

test("complete the Arabic keyboard selection, save and delete flow", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByLabel("Language", { exact: true }).selectOption("ar");
  await expect(page.getByText("وضع الاختبار.", { exact: true })).toBeVisible();
  await page
    .getByLabel("رابط الوسائط", { exact: true })
    .fill("https://example.com/sample");
  await page.getByRole("button", { name: "تحليل الرابط", exact: true }).click();
  const radios = page.getByRole("radio");
  await radios.first().focus();
  await page.keyboard.press("ArrowDown");
  await expect(page.getByRole("radio", { name: /720p/ })).toBeChecked();
  await page.getByRole("button", { name: "نزّل الملف", exact: true }).click();
  await expect(page.getByText("ملفك جاهز", { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  const pending = page.waitForEvent("download");
  await page.getByRole("link", { name: "حفظ ملف الاختبار" }).click();
  const download = await pending;
  expect(await readFile((await download.path())!, "utf8")).toContain(
    "It is not extracted media.",
  );
  await page.getByRole("button", { name: /^حذف .*A small/ }).click();
  await expect(page.getByRole("dialog")).toContainText("حذف هذا التنزيل؟");
  await page.getByRole("button", { name: "حذف التنزيل", exact: true }).click();
  await expect(page.getByText("لا توجد تنزيلات بعد")).toBeVisible();
  await expect(page.locator("#main-content")).toBeFocused();
});

test("Arabic errors, RTL reflow, local fonts and bounded keyboard dialog", async ({
  page,
}) => {
  const fontRequests: string[] = [];
  page.on("request", (request) => {
    if (request.resourceType() === "font") fontRequests.push(request.url());
  });
  await page.goto("/");
  await page.getByLabel("Language", { exact: true }).selectOption("ar");
  await page
    .getByLabel("رابط الوسائط", { exact: true })
    .fill("https://example.com/\u202ephoto");
  await page.getByRole("button", { name: "تحليل الرابط", exact: true }).click();
  await expect(page.locator("#workspace-error")).toContainText(
    "أدخل رابطًا عامًا كاملًا",
  );
  await page
    .getByLabel("رابط الوسائط", { exact: true })
    .fill("https://example.com/unsupported");
  await page.getByRole("button", { name: "تحليل الرابط", exact: true }).click();
  await expect(page.locator("#workspace-error")).toContainText(
    "لم نصل إلى وسائط عامة",
  );
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const size of [
    { width: 1440, height: 1000 },
    { width: 320, height: 740 },
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
  await page.setViewportSize({ width: 812, height: 375 });
  await page.getByRole("button", { name: "المساعدة", exact: true }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  const box = await dialog.boundingBox();
  expect(box!.y).toBeGreaterThanOrEqual(0);
  expect(box!.y + box!.height).toBeLessThanOrEqual(375);
  await page.getByRole("button", { name: "إغلاق", exact: true }).press("Enter");
  await expect(dialog).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "المساعدة", exact: true }),
  ).toBeFocused();
  await page.evaluate(async () => {
    await document.fonts.ready;
  });
  expect(fontRequests.length).toBeGreaterThan(0);
  for (const fontURL of fontRequests)
    expect(new URL(fontURL).origin).toBe(new URL(page.url()).origin);
});
