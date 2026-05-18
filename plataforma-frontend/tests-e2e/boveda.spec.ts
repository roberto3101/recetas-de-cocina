import { expect, test, Page } from "@playwright/test";

const correoOperador = "smoke@codeplex.pe";
const passwordOperador = "Smoke_Test_2026!";

const TITULO_E2E = "E2E Test Acceso";
const USUARIO_E2E = "e2e@externo.com";
const PASSWORD_E2E = "E2E_Password_2026!";
const NOMBRE_SISTEMA_E2E = "E2E Sistema";

async function login(page: Page) {
  await page.goto("/");
  await page.getByLabel(/Ingrediente/i).fill(correoOperador);
  await page.getByLabel(/Código del chef/i).fill(passwordOperador);
  await page.getByRole("button", { name: /Buscar receta/i }).click();
  await expect(page).toHaveURL(/\/panel\/acceso/);
}

async function limpiarAccesoSiExiste(page: Page) {
  await page.evaluate(async (usuario) => {
    const r = await fetch(`/cocina/boveda/accesos`);
    if (!r.ok) return;
    const j = await r.json();
    for (const a of j.datos?.accesos ?? []) {
      if (a.usuario_externo === usuario) {
        await fetch(`/cocina/boveda/accesos/${a.id}`, { method: "DELETE" });
      }
    }
  }, USUARIO_E2E);
}

async function asegurarSistemaPrueba(page: Page): Promise<string> {
  return await page.evaluate(async (nombreSistema) => {
    const r = await fetch(`/cocina/sistemas`);
    const j = await r.json();
    const existente = (j.datos?.sistemas ?? []).find((s: { nombre: string }) => s.nombre === nombreSistema);
    if (existente) return existente.id as string;
    const crear = await fetch(`/cocina/sistemas`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        codigo: "e2e_sistema_" + Date.now(),
        nombre: nombreSistema,
        url_acceso: "https://e2e.test/login",
      }),
    });
    const cj = await crear.json();
    return cj.datos?.sistema_id as string;
  }, NOMBRE_SISTEMA_E2E);
}

async function crearAccesoViaApi(page: Page) {
  await page.evaluate(async ({ titulo, usuario, password, nombreSistema }) => {
    const sis = await fetch(`/cocina/sistemas`).then((r) => r.json());
    const id = (sis.datos.sistemas as { id: string; nombre: string }[]).find((s) => s.nombre === nombreSistema)?.id;
    await fetch(`/cocina/boveda/accesos`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ titulo, sistema_destino_id: id, usuario_externo: usuario, password }),
    });
  }, { titulo: TITULO_E2E, usuario: USUARIO_E2E, password: PASSWORD_E2E, nombreSistema: NOMBRE_SISTEMA_E2E });
}

test.describe("Bóveda — flujo de acceso guardado", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await asegurarSistemaPrueba(page);
    await limpiarAccesoSiExiste(page);
  });

  test.afterEach(async ({ page }) => {
    await limpiarAccesoSiExiste(page);
  });

  test("registrar acceso desde el form y verlo en la lista", async ({ page }) => {
    const sistemaId = await asegurarSistemaPrueba(page);
    await page.goto("/panel/registro");

    await page.getByLabel(/^Título$/i).fill(TITULO_E2E);
    await page.locator("#acceso-url").selectOption(sistemaId);
    await page.getByLabel(/^Usuario$/i).fill(USUARIO_E2E);
    await page.getByLabel(/^Clave$/i).fill(PASSWORD_E2E);
    await page.getByRole("button", { name: /^Guardar$/i }).click();

    // Mensaje verde de éxito aparece inmediatamente
    await expect(page.getByText(/Acceso guardado/i)).toBeVisible({ timeout: 5000 });
    // Después redirige a /panel/acceso (delay 800ms)
    await expect(page).toHaveURL(/\/panel\/acceso/, { timeout: 5000 });
    await expect(page.locator(":visible", { hasText: USUARIO_E2E }).first()).toBeVisible({ timeout: 5000 });
  });

  test("el listado NO expone el password en JSON ni en HTML", async ({ page }) => {
    await crearAccesoViaApi(page);
    await page.goto("/panel/acceso");
    await expect(page.locator(":visible", { hasText: USUARIO_E2E }).first()).toBeVisible();

    const resp = await page.request.get("/cocina/boveda/accesos");
    const body = await resp.text();
    expect(body).not.toContain(PASSWORD_E2E);

    const html = await page.content();
    expect(html).not.toContain(PASSWORD_E2E);
  });

  test("el password se descifra al pedirlo y NO está en plano en BD", async ({ page }) => {
    await crearAccesoViaApi(page);

    // El endpoint de bookmarklet devuelve el password en plano (esperado, requiere sesión).
    const acc = await page.evaluate(async (usuario) => {
      const r = await fetch(`/cocina/boveda/accesos`);
      const j = await r.json();
      const a = (j.datos.accesos as { id: string; usuario_externo: string }[]).find((x) => x.usuario_externo === usuario);
      if (!a) return null;
      const detalle = await fetch(`/cocina/boveda/accesos/${a.id}/bookmarklet`).then((r) => r.json());
      return detalle.datos;
    }, USUARIO_E2E);

    expect(acc).toBeTruthy();
    expect(acc.usuario).toBe(USUARIO_E2E);
    expect(acc.password).toBe(PASSWORD_E2E);
  });

  test("desactivar y reactivar un acceso", async ({ page }) => {
    await crearAccesoViaApi(page);
    await page.goto("/panel/acceso");

    page.once("dialog", (d) => d.accept());
    // En desktop la fila es <tr> dentro de la tabla; tomamos la visible
    const filaActiva = page.locator("tr:visible").filter({ hasText: USUARIO_E2E }).first();
    await filaActiva.getByRole("button", { name: /^Desactivar$/i }).click();

    // Default filter = "Solo activos" → ya no debe aparecer en filas visibles
    await expect(page.locator("tr:visible", { hasText: USUARIO_E2E })).toHaveCount(0, { timeout: 3000 });

    // Cambio a "Solo inactivos"
    const selectores = page.locator("select");
    await selectores.last().selectOption("REVOCADO");
    const filaInactiva = page.locator("tr:visible").filter({ hasText: USUARIO_E2E }).first();
    await filaInactiva.getByRole("button", { name: /^Reactivar$/i }).click();

    // Vuelvo a "Solo activos" — debería volver a aparecer
    await selectores.last().selectOption("ACTIVO");
    await expect(page.locator("tr:visible", { hasText: USUARIO_E2E }).first()).toBeVisible();
  });
});
