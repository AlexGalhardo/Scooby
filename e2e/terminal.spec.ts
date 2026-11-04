import { expect, test } from "@playwright/test";

test("Google-only login, real navigation, CSV import and logout", async ({ page }) => {
  test.setTimeout(60000);
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("/");
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole("link", { name: "Continuar com Google" })).toBeVisible();
  await expect(page.getByRole("textbox")).toHaveCount(0);
  await page.getByRole("link", { name: "Continuar com Google" }).click();
  await expect(page.getByRole("heading", { name: "Visão geral" })).toBeVisible();
  await expect(page.getByText("Patrimônio simulado")).toBeVisible();
  await page.getByRole("link", { name: "Dados de mercado", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Dados de mercado" })).toBeVisible();
  await page.getByLabel("Fonte / fornecedor").fill("Synthetic E2E fixture");
  await page
    .getByLabel("Licença / autorização de uso")
    .fill("Generated test data; no market claim");
  const rows = ["time,open,high,low,close,volume"];
  const base = Date.parse("2026-10-26T13:00:00Z");
  for (let j = 0; j < 64; j++) {
    const price = 30 + [0, 0.1, 0.2, 0.1, 0, -0.1, -0.2, -0.1][j % 8];
    const open = j ? 30 + [0, 0.1, 0.2, 0.1, 0, -0.1, -0.2, -0.1][(j - 1) % 8] : 30;
    rows.push(
      `${new Date(base + j * 300000).toISOString()},${open.toFixed(2)},${(price + 0.4).toFixed(2)},${(price - 0.4).toFixed(2)},${price.toFixed(2)},100000`,
    );
  }
  await page.getByLabel("Conteúdo CSV").fill(rows.join("\n"));
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Importar e validar" }).click();
  await expect(page.getByText("64 candles importados com hash e procedência.")).toBeVisible();
  await page.getByRole("link", { name: "Estratégias", exact: true }).click();
  await page.getByRole("button", { name: "Nova estratégia" }).click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel("Nome", { exact: true }).fill("E2E moving averages");
  await dialog
    .getByLabel("Dataset para replay")
    .selectOption({ label: "PETR4 · 64 candles · Synthetic E2E fixture" });
  await dialog.getByLabel("Janela rápida").fill("2");
  await dialog.getByLabel("Janela lenta / lookback").fill("4");
  await dialog.getByRole("button", { name: "Criar configuração" }).click();
  await expect(dialog.getByText("Configuração criada e desabilitada.")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.getByRole("link", { name: "Laboratório", exact: true }).click();
  await page
    .getByLabel("Estratégia", { exact: true })
    .selectOption({ label: "E2E moving averages · PETR4" });
  await page
    .getByLabel("Dataset", { exact: true })
    .selectOption({ label: "PETR4 · 64 candles · Synthetic E2E fixture" });
  await page.getByRole("button", { name: "Enfileirar backtest" }).click();
  await expect(page.getByRole("button", { name: "Abrir", exact: true })).toBeVisible({
    timeout: 20000,
  });
  await page.getByRole("button", { name: "Abrir", exact: true }).click();
  await expect(
    page.getByRole("dialog").getByRole("button", { name: "Exportar artefato JSON" }),
  ).toBeVisible();
  await page.keyboard.press("Escape");

  for (const [label, heading] of [
    ["Estratégias", "Estratégias"],
    ["Laboratório", "Laboratório"],
    ["Ordens", "Ordens"],
    ["Posições", "Posições"],
    ["Controle de risco", "Controle de risco"],
    ["Auditoria", "Auditoria"],
    ["Configurações", "Configurações"],
  ]) {
    await page.getByRole("link", { name: label, exact: true }).click();
    await expect(page.getByRole("heading", { name: heading, exact: true })).toBeVisible();
  }
  await page.screenshot({ path: "/tmp/scooby-desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByRole("button", { name: "Toggle Sidebar" })).toBeVisible();
  await page.screenshot({ path: "/tmp/scooby-mobile.png", fullPage: true });
  expect(errors).toEqual([]);
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.getByRole("button", { name: "Sair", exact: true }).click();
  await expect(page).toHaveURL(/\/login$/);
});
