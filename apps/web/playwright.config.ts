import { defineConfig, devices } from "@playwright/test";
const port = process.env.WEB_PORT || "3000";
const origin = `http://localhost:${port}`;
export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  use: { baseURL: origin, trace: "retain-on-failure" },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    {
      name: "mobile",
      use: { ...devices["iPhone 13"], defaultBrowserType: "chromium" },
    },
  ],
  webServer: [
    {
      command: "go -C ../../services/api run ./cmd/api",
      url: "http://127.0.0.1:8080/api/v1/health",
      reuseExistingServer: !process.env.CI,
      env: {
        OD_FIXTURE_MODE: "true",
        OD_DATA_DIR: "../../.data/e2e",
        OD_ORIGIN: origin,
      },
    },
    {
      command: `pnpm dev --port ${port}`,
      url: origin,
      reuseExistingServer: !process.env.CI,
      timeout: 90_000,
      env: { NEXT_TELEMETRY_DISABLED: "1" },
    },
  ],
});
