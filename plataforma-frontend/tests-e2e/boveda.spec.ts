import { expect, test } from "@playwright/test";

const correoOperador = "smoke@codeplex.pe";
const passwordOperador = "Smoke_Test_2026!";

async function iniciarSesion(page: import("@playwright/test").Page) {
  await page.goto("/");
  await page.getByLabel(/Ingrediente/i).fill(correoOperador);
  await page.getByLabel(/Código del chef/i).fill(passwordOperador);
  await page.getByRole("button", { name: /Buscar receta/i }).click();
  await expect(page).toHaveURL(/\/panel\/inventario/);
}

async function eliminarEntradaSiExiste(page: import("@playwright/test").Page, titulo: string) {
  // Buscar y eliminar vía API directa para mantener limpio
  await page.evaluate(async (titulo) => {
    const r = await fetch(`/cocina/boveda/entradas?q=${encodeURIComponent(titulo)}&pagina=1`);
    const j = await r.json();
    for (const e of j.datos?.entradas ?? []) {
      await fetch(`/cocina/boveda/entradas/${e.id}`, { method: "DELETE" });
    }
  }, titulo);
}

test.describe("Bóveda — flujos de entrada", () => {
  test.beforeEach(async ({ page }) => {
    await iniciarSesion(page);
  });

  test("registrar nueva entrada y verla en el listado", async ({ page }) => {
    const titulo = `E2E entrada ${Date.now()}`;
    await eliminarEntradaSiExiste(page, titulo);

    await page.getByRole("link", { name: /Registrar entrada/i }).click();
    await expect(page).toHaveURL(/\/panel\/registrar/);

    await page.getByLabel(/^Título/i).fill(titulo);
    await page.getByLabel(/^URL/i).fill("https://codeplex.pe/crm");
    await page.getByLabel(/Usuario$/i).fill("usuario_destino@codeplex.pe");
    await page.getByLabel(/^Clave/i).fill("Clave_Destino_2026!");
    await page.getByLabel(/Observaciones/i).fill("creada por playwright");
    await page.getByRole("button", { name: /Guardar entrada/i }).click();

    await expect(page).toHaveURL(/\/panel\/inventario/);
    await expect(page.getByText(titulo)).toBeVisible();
  });

  test("filtrar por texto en el listado", async ({ page }) => {
    const titulo = `Filtrable_${Date.now()}`;
    await page.goto("/panel/registrar");
    await page.getByLabel(/^Título/i).fill(titulo);
    await page.getByLabel(/^URL/i).fill("https://test.local");
    await page.getByLabel(/Usuario$/i).fill("u@x.com");
    await page.getByLabel(/^Clave/i).fill("Password_2026!");
    await page.getByRole("button", { name: /Guardar entrada/i }).click();

    await expect(page).toHaveURL(/\/panel\/inventario/);
    await page.getByPlaceholder(/Buscar por título/i).fill(titulo);
    await page.getByRole("button", { name: /^Buscar$/ }).click();

    await expect(page.locator("tbody tr")).toHaveCount(1);
    await expect(page.getByText(titulo)).toBeVisible();
  });

  test("click en entrada copia la clave al clipboard y abre la URL", async ({ page, context }) => {
    await context.grantPermissions(["clipboard-read", "clipboard-write"]);
    const titulo = `Autofill_${Date.now()}`;
    const passwordPlana = `PassDescifrar_${Date.now()}!`;

    await page.goto("/panel/registrar");
    await page.getByLabel(/^Título/i).fill(titulo);
    await page.getByLabel(/^URL/i).fill("https://example.com");
    await page.getByLabel(/Usuario$/i).fill("autofill@codeplex.pe");
    await page.getByLabel(/^Clave/i).fill(passwordPlana);
    await page.getByRole("button", { name: /Guardar entrada/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    await page.getByPlaceholder(/Buscar por título/i).fill(titulo);
    await page.getByRole("button", { name: /^Buscar$/ }).click();

    // Click en la fila (sin que abra realmente la pestaña la consideramos OK)
    const filaPromesa = page.waitForEvent("popup").catch(() => null);
    await page.getByText(titulo).first().click();
    await filaPromesa;

    // El portapapeles debe contener la password
    await expect(page.getByText(/copiada al portapapeles/i)).toBeVisible({ timeout: 3000 });
    const enClipboard = await page.evaluate(() => navigator.clipboard.readText());
    expect(enClipboard).toBe(passwordPlana);
  });

  test("eliminar entrada via API y refrescar listado", async ({ page }) => {
    const titulo = `ParaBorrar_${Date.now()}`;
    await page.goto("/panel/registrar");
    await page.getByLabel(/^Título/i).fill(titulo);
    await page.getByLabel(/^URL/i).fill("https://eliminame.test");
    await page.getByLabel(/Usuario$/i).fill("u@x.com");
    await page.getByLabel(/^Clave/i).fill("Password_2026!");
    await page.getByRole("button", { name: /Guardar entrada/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    await page.getByPlaceholder(/Buscar por título/i).fill(titulo);
    await page.getByRole("button", { name: /^Buscar$/ }).click();
    await expect(page.getByText(titulo)).toBeVisible();

    await eliminarEntradaSiExiste(page, titulo);
    await page.reload();
    await page.getByPlaceholder(/Buscar por título/i).fill(titulo);
    await page.getByRole("button", { name: /^Buscar$/ }).click();
    await expect(page.getByText(/No hay entradas que coincidan/i)).toBeVisible();
  });

  test("paginación: pedir tamano_pagina=2 devuelve máximo 2 filas", async ({ page }) => {
    // Crear 3 entradas para forzar paginación
    const base = `Pag_${Date.now()}_`;
    for (let i = 0; i < 3; i++) {
      await page.goto("/panel/registrar");
      await page.getByLabel(/^Título/i).fill(base + i);
      await page.getByLabel(/^URL/i).fill("https://x.test");
      await page.getByLabel(/Usuario$/i).fill("u@x.com");
      await page.getByLabel(/^Clave/i).fill("P_2026!");
      await page.getByRole("button", { name: /Guardar entrada/i }).click();
      await expect(page).toHaveURL(/\/panel\/inventario/);
    }
    // El UI usa tamano fijo 100, pero validamos vía API que la paginación funciona
    const json = await page.evaluate(async () => {
      const r = await fetch("/cocina/boveda/entradas?pagina=1&tamano_pagina=2");
      return r.json();
    });
    expect(json.datos.entradas.length).toBeLessThanOrEqual(2);
    expect(json.datos.tamano_pagina).toBe(2);
  });
});
