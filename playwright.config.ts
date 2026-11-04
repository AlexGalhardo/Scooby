import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./e2e",
  workers: 1,
  use: {
    baseURL: "http://127.0.0.1:3100",
    browserName: "chromium",
    launchOptions: process.env.CHROMIUM_PATH
      ? { executablePath: process.env.CHROMIUM_PATH, args: ["--no-sandbox"] }
      : {},
  },
  webServer: {
    command: "go run -tags e2e ./cmd/e2e",
    url: "http://127.0.0.1:3100/up",
    reuseExistingServer: false,
    timeout: 120000,
  },
  reporter: "list",
});
