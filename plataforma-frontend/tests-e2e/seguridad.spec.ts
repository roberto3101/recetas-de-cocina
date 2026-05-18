import { expect, test, Page } from "@playwright/test";

const correoOperador = "smoke@codeplex.pe";
const passwordOperador = "Smoke_Test_2026!";

async function login(page: Page) {
  await page.goto("/");
  await page.getByLabel(/Ingrediente/i).fill(correoOperador);
  await page.getByLabel(/Código del chef/i).fill(passwordOperador);
  await page.getByRole("button", { name: /Buscar receta/i }).click();
  await expect(page).toHaveURL(/\/panel\/acceso/);
}

test.describe("Seguridad (navegador real)", () => {
  test("sin sesión, /cocina/identidad/perfil devuelve 404 sigiloso", async ({ page }) => {
    const respuesta = await page.request.get("/cocina/identidad/perfil");
    expect(respuesta.status()).toBe(404);
  });

  test("UUID malformado en /boveda/accesos devuelve 400", async ({ page }) => {
    await login(page);
    const resp = await page.request.get("/cocina/boveda/accesos/no-soy-uuid/bookmarklet");
    expect(resp.status()).toBe(400);
  });

  test("mass-assignment: campos desconocidos rechazados al crear acceso", async ({ page }) => {
    await login(page);
    const resp = await page.request.post("/cocina/boveda/accesos", {
      data: {
        titulo: "x",
        sistema_destino_id: "00000000-0000-0000-0000-000000000000",
        usuario_externo: "u@x.com",
        password: "p",
        campo_inventado: "inyectado",
      },
    });
    expect(resp.status()).toBe(400);
  });

  test("la SPA no expone secretos en el HTML inicial", async ({ page }) => {
    await page.goto("/");
    const html = await page.content();
    expect(html).not.toContain("KEK");
    expect(html).not.toContain("PGPASSWORD");
    expect(html).not.toContain("API_KEY");
    expect(html).not.toContain("BASE_DATOS_URL");
  });

  test("la cookie de sesión es HttpOnly, Secure y SameSite=Strict", async ({ page, context }) => {
    await login(page);
    const cookies = await context.cookies();
    const sesion = cookies.find((c) => c.name === "sesion_cocina");
    expect(sesion).toBeDefined();
    if (sesion) {
      expect(sesion.httpOnly).toBe(true);
      expect(sesion.sameSite).toBe("Strict");
      // En producción HTTPS Secure debe ser true; en dev local HTTP puede ser false.
      // Solo afirmar Secure si el origin es https.
      if (page.url().startsWith("https://")) {
        expect(sesion.secure).toBe(true);
      }
    }
  });

  test("el icon SVG está disponible (PWA básica)", async ({ page }) => {
    const resp = await page.request.get("/icon.svg");
    expect(resp.status()).toBe(200);
    const ct = resp.headers()["content-type"] ?? "";
    expect(ct).toContain("svg");
  });
});
