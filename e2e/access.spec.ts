import { expect, test } from "@playwright/test";

const pages = [
  "/",
  "/market",
  "/strategies",
  "/backtests",
  "/orders",
  "/positions",
  "/risk",
  "/audit",
  "/settings",
  "/dashboard",
];
const protectedGets = [
  "/api/snapshot",
  "/api/strategy-kinds",
  "/api/metrics",
  "/api/stream",
  "/api/preferences",
  "/api/artifacts/none",
  "/api/datasets/none",
  "/api/backtests/none",
];
const protectedWrites: [string, string][] = [
  ["PUT", "/api/preferences"],
  ["POST", "/api/datasets/import"],
  ["POST", "/api/datasets/fetch"],
  ["POST", "/api/strategies"],
  ["POST", "/api/strategies/draft"],
  ["POST", "/api/strategies/none/toggle"],
  ["POST", "/api/strategies/none/decision"],
  ["POST", "/api/backtests"],
  ["POST", "/api/backtests/none/cancel"],
  ["POST", "/api/accounts/b3/policy"],
  ["POST", "/api/accounts/b3/ticket"],
  ["POST", "/api/accounts/b3/control"],
  ["POST", "/api/accounts/b3/orders/none/cancel"],
];

test("anonymous visitors only reach the login page and the health endpoints", async ({
  page,
  request,
}) => {
  for (const path of pages) {
    await page.goto(path);
    await expect(page, path).toHaveURL(/\/login$/);
  }
  for (const path of ["/ready", "/up"]) {
    expect((await request.get(path)).status(), path).toBe(200);
  }
});

test("every private API answers 401 without a session", async ({ playwright, baseURL }) => {
  const request = await playwright.request.newContext({
    baseURL,
    extraHTTPHeaders: { Accept: "application/json" },
  });
  const statuses: Record<string, number> = {};
  for (const path of protectedGets) statuses[path] = (await request.get(path)).status();
  expect(statuses).toEqual(Object.fromEntries(protectedGets.map((p) => [p, 401])));
  for (const [method, path] of protectedWrites) {
    const response = await request.fetch(path, { method, data: {} });
    expect([401, 419], `${method} ${path}`).toContain(response.status());
  }
  expect((await request.post("/logout", { maxRedirects: 0 })).status()).not.toBe(200);
  await request.dispose();
});

test("the session cookie is HttpOnly and the private API works once signed in", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("link", { name: "Continue with Google" }).click();
  await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();
  const cookies = await page.context().cookies();
  expect(cookies.length).toBeGreaterThan(0);
  expect(cookies.every((c) => c.httpOnly || c.name.toLowerCase().includes("xsrf"))).toBe(true);
  for (const path of ["/api/snapshot", "/api/strategy-kinds", "/api/preferences"]) {
    expect((await page.request.get(path)).status(), path).toBe(200);
  }
  expect((await page.request.get("/api/artifacts/none")).status()).toBe(404);
  await page.goto("/login");
  await expect(page).toHaveURL(/\/$/);
});

test("a signed-in owner can chart imported history with indicators", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("link", { name: "Continue with Google" }).click();
  await page.getByRole("link", { name: "Market data", exact: true }).click();
  await page.getByLabel("Source / vendor").fill("Synthetic indicator fixture");
  await page.getByLabel("License / authorized use").fill("Generated test data; no market claim");
  const rows = ["time,open,high,low,close,volume"];
  const base = Date.parse("2026-10-26T13:00:00Z");
  for (let j = 0; j < 80; j++) {
    const close = 20 + Math.sin(j / 5) * 2;
    rows.push(
      `${new Date(base + j * 300000).toISOString()},${close.toFixed(2)},${(close + 0.3).toFixed(2)},${(close - 0.3).toFixed(2)},${close.toFixed(2)},1000`,
    );
  }
  await page.getByLabel("CSV content").fill(rows.join("\n"));
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Import and validate" }).click();
  await expect(page.getByText("80 candles imported with hash and provenance.")).toBeVisible();
  await page.getByRole("button", { name: "Chart" }).first().click();
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  const chart = page.getByRole("img", { name: /Market chart/ });
  await expect(chart).toBeVisible();
  for (const label of ["SMA 20", "EMA 50", "Bollinger 20", "VWAP", "Volume", "RSI 14"]) {
    await page.getByLabel(label, { exact: true }).check();
    await expect(page.getByLabel(label, { exact: true })).toBeChecked();
  }
  await expect(chart).toBeVisible();
  await page.getByLabel("RSI 14", { exact: true }).uncheck();
  await expect(page.getByLabel("RSI 14", { exact: true })).not.toBeChecked();
  expect(errors).toEqual([]);
});
