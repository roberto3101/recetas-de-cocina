import { expect, test } from "@playwright/test";

const correoOperador = "smoke@codeplex.pe";
const passwordOperador = "Smoke_Test_2026!";

test.describe("Flujo de identidad (cebo + sesión)", () => {
  test("la página de inicio se muestra como blog de recetas", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveTitle(/Recetas del Chef/);
    await expect(page.getByRole("heading", { name: "Recetas del Chef" })).toBeVisible();
    await expect(page.getByPlaceholder(/papa amarilla|ingrediente/i)).toBeVisible();
    await expect(page.getByRole("button", { name: /Buscar receta/i })).toBeVisible();
  });

  test("credenciales inválidas no revelan que es un login", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill("noexiste@codeplex.pe");
    await page.getByLabel(/Código del chef/i).fill("clave-mala");
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/sin-resultados/);
    await expect(page.getByRole("heading", { name: /Sin resultados/i })).toBeVisible();
    // No debe filtrar palabras tipo "login", "credenciales" en la página
    const cuerpo = await page.content();
    expect(cuerpo.toLowerCase()).not.toContain("credenciales");
    expect(cuerpo.toLowerCase()).not.toContain("login");
    expect(cuerpo.toLowerCase()).not.toContain("password");
  });

  test("login con credenciales válidas redirige al panel", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);
    await expect(page.getByRole("heading", { name: /Inventario/i }).first()).toBeVisible();
    await expect(page.locator("header").getByText(correoOperador)).toBeVisible();
  });

  test("acceder /panel/inventario sin sesión redirige al cebo", async ({ page }) => {
    await page.goto("/panel/inventario");
    await expect(page).toHaveURL("/");
  });

  test("la cookie de sesión NO es accesible desde JavaScript (HttpOnly)", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    const cookiesJsVisibles = await page.evaluate(() => document.cookie);
    expect(cookiesJsVisibles).not.toContain("sesion_cocina");
  });

  test("cerrar sesión vuelve al cebo y revoca la sesión", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    await page.getByRole("button", { name: /Cerrar sesión/i }).click();
    await expect(page).toHaveURL("/");

    // Intentar volver al panel debe redirigir al inicio
    await page.goto("/panel/inventario");
    await expect(page).toHaveURL("/");
  });
});
