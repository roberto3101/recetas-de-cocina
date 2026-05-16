import { expect, test } from "@playwright/test";

const correoOperador = "smoke@codeplex.pe";
const passwordOperador = "Smoke_Test_2026!";

test.describe("Seguridad (navegador real)", () => {
  test("no se pueden ver datos de operador desde JS sin sesión", async ({ page }) => {
    const respuesta = await page.request.get("/cocina/identidad/perfil");
    expect(respuesta.status()).toBe(404);
  });

  test("payload XSS guardado se renderiza como texto, no como HTML", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    const tituloPayload = `<img src=x onerror="window.__XSS_EJECUTADO=true">`;
    await page.goto("/panel/registrar");
    await page.getByLabel(/^Título/i).fill(tituloPayload);
    await page.getByLabel(/^URL/i).fill("https://example.com");
    await page.getByLabel(/Usuario$/i).fill("xss@codeplex.pe");
    await page.getByLabel(/^Clave/i).fill("Password_2026!");
    await page.getByRole("button", { name: /Guardar entrada/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    // El título debe verse como texto literal, no ejecutarse
    await page.waitForTimeout(500);
    const seEjecuto = await page.evaluate(() => (window as unknown as { __XSS_EJECUTADO?: boolean }).__XSS_EJECUTADO === true);
    expect(seEjecuto).toBe(false);

    // Debe aparecer como texto en el listado (puede haber duplicados de runs previos)
    await expect(page.getByText(tituloPayload, { exact: true }).first()).toBeVisible();
  });

  test("intentar login 6 veces con password mal bloquea al operador", async ({ page, context }) => {
    // Crear un operador efímero vía API admin sería ideal; aquí usamos el SMOKE existente
    // y verificamos que tras varios intentos malos no podemos ingresar con la buena.
    // (Este test es DESTRUCTIVO sobre el operador smoke; lo dejamos comentado por default.)
    test.skip(true, "test destructivo - bloquearía al operador smoke en BD local");
    await page.goto("/");
    for (let i = 0; i < 6; i++) {
      await page.getByLabel(/Ingrediente/i).fill(correoOperador);
      await page.getByLabel(/Código del chef/i).fill("password-incorrecto");
      await page.getByRole("button", { name: /Buscar receta/i }).click();
      await expect(page).toHaveURL(/\/sin-resultados/);
      await page.goto("/");
    }
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/sin-resultados/);
    void context;
  });

  test("URL malformada con UUID inválido se rechaza con 400", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    const resp = await page.request.get("/cocina/boveda/entradas/no-soy-uuid");
    expect(resp.status()).toBe(400);
  });

  test("mass-assignment: campos desconocidos rechazados", async ({ page }) => {
    await page.goto("/");
    await page.getByLabel(/Ingrediente/i).fill(correoOperador);
    await page.getByLabel(/Código del chef/i).fill(passwordOperador);
    await page.getByRole("button", { name: /Buscar receta/i }).click();
    await expect(page).toHaveURL(/\/panel\/inventario/);

    const resp = await page.request.post("/cocina/boveda/entradas", {
      data: {
        titulo: "x",
        tipo: "WEB",
        usuario_acceso: "u@x.com",
        password_acceso: "p",
        campo_inventado: "inyectado",
      },
    });
    expect(resp.status()).toBe(400);
  });

  test("la SPA no expone secretos en el HTML inicial", async ({ page }) => {
    await page.goto("/");
    const html = await page.content();
    // Heurística: no debe haber strings como "KEK", "PGPASSWORD", "API_KEY"
    expect(html).not.toContain("KEK");
    expect(html).not.toContain("PGPASSWORD");
    expect(html).not.toContain("API_KEY");
    expect(html).not.toContain("BASE_DATOS_URL");
  });
});
