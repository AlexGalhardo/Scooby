import { expect, test } from "@playwright/test";

test("Google-only login, real navigation, CSV import and logout", async ({ page }) => {
  test.setTimeout(60000);
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("/");
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole("link", { name: "Continue with Google" })).toBeVisible();
  await expect(page.getByRole("textbox")).toHaveCount(0);
  await page.getByRole("link", { name: "Continue with Google" }).click();
  await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();
  await expect(page.getByText("Simulated equity")).toBeVisible();
  await page.getByRole("link", { name: "Market data", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Market data" })).toBeVisible();
  await page.getByLabel("Source / vendor").fill("Synthetic E2E fixture");
  await page.getByLabel("License / authorized use").fill("Generated test data; no market claim");
  const rows = ["time,open,high,low,close,volume"];
  const base = Date.parse("2026-10-26T13:00:00Z");
  for (let j = 0; j < 64; j++) {
    const price = 30 + [0, 0.1, 0.2, 0.1, 0, -0.1, -0.2, -0.1][j % 8];
    const open = j ? 30 + [0, 0.1, 0.2, 0.1, 0, -0.1, -0.2, -0.1][(j - 1) % 8] : 30;
    rows.push(
      `${new Date(base + j * 300000).toISOString()},${open.toFixed(2)},${(price + 0.4).toFixed(2)},${(price - 0.4).toFixed(2)},${price.toFixed(2)},100000`,
    );
  }
  await page.getByLabel("CSV content").fill(rows.join("\n"));
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Import and validate" }).click();
  await expect(page.getByText("64 candles imported with hash and provenance.")).toBeVisible();
  await page.getByRole("link", { name: "Strategies", exact: true }).click();
  await page.getByRole("button", { name: "New strategy" }).click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel("Name", { exact: true }).fill("E2E moving averages");
  await dialog
    .getByLabel("Dataset for replay")
    .selectOption({ label: "PETR4 · 64 candles · Synthetic E2E fixture" });
  await dialog.getByLabel("Fast window").fill("2");
  await dialog.getByLabel("Slow window").fill("4");
  await dialog.getByRole("button", { name: "Create configuration" }).click();
  await expect(dialog.getByText("Configuration created and disabled.")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.getByRole("link", { name: "Lab", exact: true }).click();
  await page
    .getByLabel("Strategy", { exact: true })
    .selectOption({ label: "E2E moving averages · PETR4" });
  await page
    .getByLabel("Dataset", { exact: true })
    .selectOption({ label: "PETR4 · 64 candles · Synthetic E2E fixture" });
  await page.getByRole("button", { name: "Enqueue backtest" }).click();
  await expect(page.getByRole("button", { name: "Open", exact: true })).toBeVisible({
    timeout: 20000,
  });
  await page.getByRole("button", { name: "Open", exact: true }).click();
  await expect(
    page.getByRole("dialog").getByRole("button", { name: "Export JSON artifact" }),
  ).toBeVisible();
  await page.keyboard.press("Escape");

  for (const [label, heading] of [
    ["Strategies", "Strategies"],
    ["Lab", "Lab"],
    ["Orders", "Orders"],
    ["Positions", "Positions"],
    ["Risk control", "Risk control"],
    ["Audit", "Audit"],
    ["Settings", "Settings"],
  ]) {
    await page.getByRole("link", { name: label, exact: true }).click();
    await expect(page.getByRole("heading", { name: heading, exact: true })).toBeVisible();
  }
  await page.screenshot({ path: "/tmp/vukem-trading-desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByRole("button", { name: "Toggle Sidebar" })).toBeVisible();
  await page.screenshot({ path: "/tmp/vukem-trading-mobile.png", fullPage: true });
  expect(errors).toEqual([]);
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await expect(page).toHaveURL(/\/login$/);
});

async function signIn(page: import("@playwright/test").Page) {
  await page.goto("/");
  await page.getByRole("link", { name: "Continue with Google" }).click();
  await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();
}

test("command palette and numbered shortcuts move between pages", async ({ page }) => {
  await signIn(page);
  await page.keyboard.press("Control+k");
  await page.getByLabel("Search commands").fill("risk");
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(/\/risk$/);
  await expect(page.getByRole("heading", { name: "Risk control", exact: true })).toBeVisible();
  await page.keyboard.press("Alt+1");
  await expect(page).toHaveURL(/\/$/);
});

test("the workspace is saved on the server and survives a reload", async ({ page }) => {
  await signIn(page);
  const saved = () =>
    page.waitForResponse(
      (r) => r.url().endsWith("/api/preferences") && r.request().method() === "PUT",
    );
  const account = page.getByLabel("Account", { exact: true });
  let wait = saved();
  await account.selectOption("us");
  await wait;
  await page.reload();
  await expect(page.getByLabel("Account", { exact: true })).toHaveValue("us");
  wait = saved();
  await page.getByLabel("Account", { exact: true }).selectOption("b3");
  await wait;
});

test("the safety bar states the account status and ticket hotkeys call the risk engine", async ({
  page,
}) => {
  await signIn(page);
  await expect(page.getByRole("region", { name: "Safety status" })).toContainText("Blocked");
  const bodies: string[] = [];
  await page.route("**/api/accounts/b3/ticket", async (route) => {
    bodies.push(route.request().postData() ?? "");
    await route.fulfill({ status: 201, contentType: "application/json", body: "{}" });
  });
  await page.getByRole("heading", { name: "Overview" }).click();
  await page.keyboard.press("b");
  await expect.poll(() => bodies.length).toBe(1);
  expect(JSON.parse(bodies[0])).toEqual({ key: "b", symbol: "PETR4" });
  await page.keyboard.press("Shift+B");
  await expect.poll(() => bodies.length).toBe(2);
  expect(JSON.parse(bodies[1]).key).toBe("shift+b");
});
