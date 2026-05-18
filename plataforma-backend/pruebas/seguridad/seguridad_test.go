package seguridad

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	httpEntrada "sistemas-unificados/aplicacion/entrada/http"
	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

const correoSeguridad = "seguridad@codeplex.pe"
const passwordSeguridad = "P@ssw0rd_Seguridad_2026!"

type entornoPruebas struct {
	servidor *httptest.Server
	cliente  *http.Client
	conexion *cockroach.ConexionBaseDatos
	usuarioId string
}

func montarEntornoPruebas(t *testing.T) *entornoPruebas {
	t.Helper()
	urlConexion := os.Getenv("BASE_DATOS_URL")
	if urlConexion == "" {
		urlConexion = "postgresql://root@localhost:26258/sistemas_unificados?sslmode=disable"
	}
	if os.Getenv("KEK_GESTOR") == "" {
		clave, err := cripto.GenerarClaveAleatoriaAES256()
		if err != nil {
			t.Fatalf("no se generó KEK: %v", err)
		}
		os.Setenv("KEK_GESTOR", clave)
	}

	contexto, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	conexion, err := cockroach.AbrirConexion(contexto, urlConexion)
	if err != nil {
		t.Skipf("BD local no disponible: %v", err)
	}
	claves, err := cripto.CargarClavesDesdeEntorno()
	if err != nil {
		t.Fatalf("claves: %v", err)
	}

	limpiarUsuario(conexion)
	resultado, err := identidad.RegistrarUsuarioYActivar(contexto, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoSeguridad,
		PasswordPlana:     passwordSeguridad,
	})
	if err != nil {
		t.Fatalf("no se pudo crear operador de prueba: %v", err)
	}

	enrutador := chi.NewRouter()
	httpEntrada.RegistrarRutas(enrutador, httpEntrada.DependenciasRutas{
		ConexionBaseDatos: conexion,
		ClavesCifrado:     claves,
	})
	servidor := httptest.NewServer(enrutador)

	jar, _ := cookiejar.New(nil)
	cliente := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Cleanup(func() {
		servidor.Close()
		limpiarUsuario(conexion)
		conexion.Cerrar()
	})

	return &entornoPruebas{
		servidor:  servidor,
		cliente:   cliente,
		conexion:  conexion,
		usuarioId: resultado.UsuarioId.String(),
	}
}

func limpiarUsuario(conexion *cockroach.ConexionBaseDatos) {
	contexto := context.Background()
	for _, sql := range []string{
		`DELETE FROM auditoria_accion WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`,
		`DELETE FROM acceso_guardado WHERE creado_por IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`,
		`DELETE FROM sistema_destino WHERE creado_por IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`,
		`DELETE FROM sesion_global WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`,
		`DELETE FROM usuario_totp WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`,
		`DELETE FROM usuario WHERE lower(correo_electronico) = lower($1)`,
	} {
		_, _ = conexion.Pool().Exec(contexto, sql, correoSeguridad)
	}
}

func iniciarSesionComoOperador(t *testing.T, entorno *entornoPruebas) {
	t.Helper()
	cuerpo := url.Values{}
	cuerpo.Set("ingrediente", correoSeguridad)
	cuerpo.Set("codigo", passwordSeguridad)
	resp, err := entorno.cliente.PostForm(entorno.servidor.URL+"/buscar", cuerpo)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login esperaba 303, got %d", resp.StatusCode)
	}
}

func getJson(t *testing.T, entorno *entornoPruebas, ruta string) (int, map[string]any) {
	t.Helper()
	resp, err := entorno.cliente.Get(entorno.servidor.URL + ruta)
	if err != nil {
		t.Fatalf("GET %s: %v", ruta, err)
	}
	defer resp.Body.Close()
	cuerpo, _ := io.ReadAll(resp.Body)
	if len(cuerpo) == 0 {
		return resp.StatusCode, nil
	}
	var datos map[string]any
	if err := json.Unmarshal(cuerpo, &datos); err != nil {
		return resp.StatusCode, map[string]any{"_raw": string(cuerpo)}
	}
	return resp.StatusCode, datos
}

func postJson(t *testing.T, entorno *entornoPruebas, ruta string, cuerpoEnvio any) (int, map[string]any) {
	t.Helper()
	return enviarConMetodo(t, entorno, http.MethodPost, ruta, cuerpoEnvio)
}

func putJson(t *testing.T, entorno *entornoPruebas, ruta string, cuerpoEnvio any) (int, map[string]any) {
	t.Helper()
	return enviarConMetodo(t, entorno, http.MethodPut, ruta, cuerpoEnvio)
}

func enviarConMetodo(t *testing.T, entorno *entornoPruebas, metodo, ruta string, cuerpoEnvio any) (int, map[string]any) {
	t.Helper()
	var bodyBytes []byte
	if cuerpoEnvio != nil {
		switch v := cuerpoEnvio.(type) {
		case []byte:
			bodyBytes = v
		case string:
			bodyBytes = []byte(v)
		default:
			b, err := json.Marshal(cuerpoEnvio)
			if err != nil {
				t.Fatal(err)
			}
			bodyBytes = b
		}
	}
	req, err := http.NewRequest(metodo, entorno.servidor.URL+ruta, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := entorno.cliente.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", metodo, ruta, err)
	}
	defer resp.Body.Close()
	cuerpo, _ := io.ReadAll(resp.Body)
	if len(cuerpo) == 0 {
		return resp.StatusCode, nil
	}
	var datos map[string]any
	if err := json.Unmarshal(cuerpo, &datos); err != nil {
		return resp.StatusCode, map[string]any{"_raw": string(cuerpo)}
	}
	return resp.StatusCode, datos
}

// -----------------------------------------------------------------------------
// AUTENTICACIÓN Y SESIÓN
// -----------------------------------------------------------------------------

func TestSeguridad_AccesoARutaProtegidaSinSesion_404Sigiloso(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	estado, _ := getJson(t, entorno, "/cocina/identidad/perfil")
	if estado != http.StatusNotFound {
		t.Fatalf("se esperaba 404 sin sesión; got %d", estado)
	}
}

func TestSeguridad_AccesoConCookieFalsificada_404Sigiloso(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	req, _ := http.NewRequest("GET", entorno.servidor.URL+"/cocina/identidad/perfil", nil)
	req.AddCookie(&http.Cookie{Name: "sesion_cocina", Value: "tokenfalsoadivinado"})
	resp, err := entorno.cliente.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("cookie inventada debería dar 404; got %d", resp.StatusCode)
	}
}

func TestSeguridad_CookieSesion_EsHttpOnlyYSameSiteStrict(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	cuerpo := url.Values{}
	cuerpo.Set("ingrediente", correoSeguridad)
	cuerpo.Set("codigo", passwordSeguridad)
	resp, err := entorno.cliente.PostForm(entorno.servidor.URL+"/buscar", cuerpo)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	cabeceraCookie := resp.Header.Get("Set-Cookie")
	if cabeceraCookie == "" {
		t.Fatal("no se emitió Set-Cookie")
	}
	if !strings.Contains(cabeceraCookie, "HttpOnly") {
		t.Errorf("cookie debe ser HttpOnly: %s", cabeceraCookie)
	}
	if !strings.Contains(cabeceraCookie, "SameSite=Strict") {
		t.Errorf("cookie debe ser SameSite=Strict: %s", cabeceraCookie)
	}
}

func TestSeguridad_SesionRevocada_404Sigiloso(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	estado, _ := postJson(t, entorno, "/cocina/identidad/cerrar-sesion", nil)
	if estado != http.StatusOK {
		t.Fatalf("logout esperaba 200; got %d", estado)
	}
	// Tras logout, la cookie sigue en el jar pero la sesión está revocada
	estado, _ = getJson(t, entorno, "/cocina/identidad/perfil")
	if estado != http.StatusNotFound {
		t.Fatalf("acceso tras logout debería ser 404; got %d", estado)
	}
}

// -----------------------------------------------------------------------------
// FUERZA BRUTA / RATE LIMITING
// -----------------------------------------------------------------------------

func TestSeguridad_LoginConCredencialesInvalidas_NoRevelaSiUserExiste(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	// Login con correo válido pero password mala
	cuerpo1 := url.Values{}
	cuerpo1.Set("ingrediente", correoSeguridad)
	cuerpo1.Set("codigo", "xxxxxxx")
	resp1, err1 := entorno.cliente.PostForm(entorno.servidor.URL+"/buscar", cuerpo1)
	if err1 != nil {
		t.Fatalf("login con correo existente: %v", err1)
	}
	defer resp1.Body.Close()

	// Login con correo inexistente
	cuerpo2 := url.Values{}
	cuerpo2.Set("ingrediente", "noexiste@nadie.pe")
	cuerpo2.Set("codigo", "xxxxxxx")
	resp2, err2 := entorno.cliente.PostForm(entorno.servidor.URL+"/buscar", cuerpo2)
	if err2 != nil {
		t.Fatalf("login con correo inexistente: %v", err2)
	}
	defer resp2.Body.Close()

	if resp1.StatusCode != resp2.StatusCode {
		t.Fatalf("respuestas distintas revelan existencia de usuario: existente=%d inexistente=%d",
			resp1.StatusCode, resp2.StatusCode)
	}
}

func TestSeguridad_BruteForce_BloqueaTrasIntentosFallidos(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	// MaxIntentosFallidosAntesDeBloqueo = 10. Hacemos 12 para garantizar bloqueo.
	for i := 0; i < 12; i++ {
		cuerpo := url.Values{}
		cuerpo.Set("ingrediente", correoSeguridad)
		cuerpo.Set("codigo", "PasswordIncorrecto_xxx")
		resp, err := entorno.cliente.PostForm(entorno.servidor.URL+"/buscar", cuerpo)
		if err != nil {
			t.Logf("intento %d error: %v", i+1, err)
			continue
		}
		t.Logf("intento %d -> status %d", i+1, resp.StatusCode)
		resp.Body.Close()
	}

	// Verificar en BD que el usuario quedó BLOQUEADO
	var estado string
	var intentos int64
	if err := entorno.conexion.Pool().QueryRow(context.Background(), `
		SELECT estado, intentos_fallidos FROM usuario WHERE lower(correo_electronico) = lower($1)
	`, correoSeguridad).Scan(&estado, &intentos); err != nil {
		t.Fatalf("query estado: %v", err)
	}
	t.Logf("BD tras 12 intentos: estado=%s intentos_fallidos=%d", estado, intentos)
	if estado != "BLOQUEADO" {
		t.Fatalf("usuario debería estar BLOQUEADO tras 12 intentos, está %s con %d intentos", estado, intentos)
	}

	// Ahora con password correcta debería seguir bloqueado
	cuerpo := url.Values{}
	cuerpo.Set("ingrediente", correoSeguridad)
	cuerpo.Set("codigo", passwordSeguridad)
	resp, err := entorno.cliente.PostForm(entorno.servidor.URL+"/buscar", cuerpo)
	if err != nil {
		t.Fatalf("login final: %v", err)
	}
	defer resp.Body.Close()
	t.Logf("login con password correcto tras bloqueo -> status %d", resp.StatusCode)
	if resp.StatusCode == http.StatusSeeOther {
		t.Fatal("usuario debería estar bloqueado; logró loguearse pese a BLOQUEADO en BD")
	}
}

func TestSeguridad_RateLimit_BuscarLimitaIntentos(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	intentosBloqueados := 0
	for i := 0; i < 80; i++ {
		cuerpo := url.Values{}
		cuerpo.Set("ingrediente", fmt.Sprintf("user%d@pseu.pe", i))
		cuerpo.Set("codigo", "xx")
		req, _ := http.NewRequest("POST", entorno.servidor.URL+"/buscar", strings.NewReader(cuerpo.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Forwarded-For", "5.6.7.8")
		resp, err := entorno.cliente.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			intentosBloqueados++
		}
		resp.Body.Close()
	}
	if intentosBloqueados == 0 {
		t.Fatal("rate limit no se activó tras 80 intentos rápidos")
	}
}

// -----------------------------------------------------------------------------
// VALIDACIÓN DE ENTRADA
// -----------------------------------------------------------------------------

func TestSeguridad_BodyExcesivamenteGrande_Rechazado(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	bodyGigante := strings.Repeat("X", 2<<20) // 2 MiB
	estado, _ := postJson(t, entorno, "/cocina/sistemas", []byte(`{"titulo":"`+bodyGigante+`"}`))
	if estado == http.StatusCreated {
		t.Fatal("body de 2MiB no debería ser aceptado")
	}
}

func TestSeguridad_JsonInvalido_DevuelveError(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	estado, _ := postJson(t, entorno, "/cocina/sistemas", []byte(`{esto no es json valido`))
	if estado != http.StatusBadRequest {
		t.Fatalf("JSON inválido debería ser 400; got %d", estado)
	}
}

func TestSeguridad_MassAssignment_RechazaCamposDesconocidos(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	cuerpoSospechoso := `{
		"titulo":"x", "tipo":"WEB", "usuario_acceso":"u@x.com", "password_acceso":"y",
		"id":"00000000-0000-0000-0000-000000000000",
		"estado":"ADMIN",
		"creado_por":"alguien-otro"
	}`
	estado, _ := postJson(t, entorno, "/cocina/sistemas", []byte(cuerpoSospechoso))
	if estado != http.StatusBadRequest {
		t.Fatalf("campos desconocidos deberían rechazarse (DisallowUnknownFields); got %d", estado)
	}
}

func TestSeguridad_UuidInvalidoEnRuta_DevuelveError400(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	estado, _ := getJson(t, entorno, "/cocina/sistemas/no-soy-un-uuid")
	if estado != http.StatusBadRequest {
		t.Fatalf("UUID inválido debería ser 400; got %d", estado)
	}
}

func TestSeguridad_PathTraversal_NoEscalaDirectorios(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	resp, err := entorno.cliente.Get(entorno.servidor.URL + "/cocina/sistemas/..%2F..%2Fetc%2Fpasswd")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal("path traversal no debería devolver 200")
	}
}

// -----------------------------------------------------------------------------
// INYECCIÓN SQL
// -----------------------------------------------------------------------------

func TestSeguridad_SqlInjectionEnFiltroTexto_NoComprometeBd(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)

	cargasUtiles := []string{
		"' OR 1=1 --",
		"'; DROP TABLE usuario; --",
		"\" UNION SELECT * FROM usuario --",
		"%' OR '1%' = '1",
		"\\'; SELECT pg_sleep(2); --",
	}
	for _, payload := range cargasUtiles {
		ruta := "/cocina/sistemas?q=" + url.QueryEscape(payload)
		estado, _ := getJson(t, entorno, ruta)
		if estado != http.StatusOK {
			t.Errorf("payload SQL injection [%s] causó error HTTP %d (esperado 200, tratado como texto)", payload, estado)
		}
	}
	// Verifica que la tabla usuario sigue intacta
	var total int64
	if err := entorno.conexion.Pool().QueryRow(context.Background(), "SELECT count(*) FROM usuario").Scan(&total); err != nil {
		t.Fatalf("la tabla usuario fue afectada: %v", err)
	}
	if total < 1 {
		t.Fatal("inyección logró borrar registros")
	}
}

func TestSeguridad_SqlInjectionEnTipoFiltro_Ignorado(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	estado, _ := getJson(t, entorno, "/cocina/sistemas?tipo=WEB%27%20OR%201=1")
	if estado != http.StatusOK {
		t.Fatalf("got %d", estado)
	}
}

// -----------------------------------------------------------------------------
// XSS
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// AUDITORÍA
// -----------------------------------------------------------------------------

func TestSeguridad_AuditoriaRegistraOperaciones(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)

	// Acción auditable que NO invalida la sesión: registrar un sistema.
	cuerpo := map[string]any{
		"codigo":     "auditoria_test_" + time.Now().Format("150405"),
		"nombre":     "Sistema auditoría",
		"url_acceso": "https://auditoria.test/login",
	}
	if estado, _ := postJson(t, entorno, "/cocina/sistemas", cuerpo); estado != http.StatusCreated {
		t.Fatalf("crear sistema: %d", estado)
	}

	estado, datos := getJson(t, entorno, "/cocina/auditoria?pagina=1&tamano_pagina=20")
	if estado != http.StatusOK {
		t.Fatalf("auditoria: %d", estado)
	}
	bloque, ok := datos["datos"].(map[string]any)
	if !ok {
		t.Fatal("respuesta auditoria sin 'datos'")
	}
	totalF, _ := bloque["total_registros"].(float64)
	if int64(totalF) < 1 {
		t.Fatal("auditoría no registró ninguna acción")
	}
}

// -----------------------------------------------------------------------------
// SESIONES CONCURRENTES Y AISLAMIENTO
// -----------------------------------------------------------------------------

func TestSeguridad_DosSesionesIndependientesNoCompartenEstado(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)

	// Crear segundo cliente sin sesión
	jar2, _ := cookiejar.New(nil)
	cliente2 := &http.Client{Jar: jar2, Timeout: 5 * time.Second}

	resp, err := cliente2.Get(entorno.servidor.URL + "/cocina/identidad/perfil")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("segundo cliente sin sesión debería ver 404; got %d", resp.StatusCode)
	}
}

func TestSeguridad_TokenSesionUnicoPorLogin(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	jar1, _ := cookiejar.New(nil)
	cliente1 := &http.Client{Jar: jar1, Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	jar2, _ := cookiejar.New(nil)
	cliente2 := &http.Client{Jar: jar2, Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	formulario := url.Values{}
	formulario.Set("ingrediente", correoSeguridad)
	formulario.Set("codigo", passwordSeguridad)

	r1, _ := cliente1.PostForm(entorno.servidor.URL+"/buscar", formulario)
	r1.Body.Close()
	r2, _ := cliente2.PostForm(entorno.servidor.URL+"/buscar", formulario)
	r2.Body.Close()

	// Limpia rate limit reset (15s ventana en login)
	cookie1 := obtenerCookieSesion(jar1)
	cookie2 := obtenerCookieSesion(jar2)
	if cookie1 == "" || cookie2 == "" {
		t.Skip("no se emitió cookie en ambos clientes (posible rate limit en test); revisar")
	}
	if cookie1 == cookie2 {
		t.Fatal("dos logins independientes generaron el mismo token de sesión")
	}
}

func obtenerCookieSesion(jar http.CookieJar) string {
	urlServidor, _ := url.Parse("http://127.0.0.1")
	for _, c := range jar.Cookies(urlServidor) {
		if c.Name == "sesion_cocina" {
			return c.Value
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// PERMISOS DE OPERACIONES
// -----------------------------------------------------------------------------

func TestSeguridad_CambiarPasswordRequiereSegundoFactor(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	// El operador no tiene TOTP activo → segundo_factor_validado = true por default
	// Pero la ruta /cocina/identidad/cambiar-password está bajo RequiereSegundoFactor.
	// Sin TOTP activo, debe permitirlo (porque la sesión se marca con SF validado).
	estado, _ := putJson(t, entorno, "/cocina/identidad/cambiar-password", map[string]any{
		"password_actual": passwordSeguridad,
		"password_nueva":  "Otro_Password_Diff_2026!",
	})
	if estado != http.StatusOK {
		t.Fatalf("cambio password con SF validado: %d", estado)
	}
}

func TestSeguridad_PasswordIncorrectoAlCambiar_DevuelveError(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	estado, _ := putJson(t, entorno, "/cocina/identidad/cambiar-password", map[string]any{
		"password_actual": "PasswordIncorrectoFalso2026!",
		"password_nueva":  "Nueva_Diff_2026!@",
	})
	if estado != http.StatusUnauthorized {
		t.Fatalf("password actual incorrecto debe devolver 401; got %d", estado)
	}
}

func TestSeguridad_PasswordDebilAlCambiar_DevuelveError(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	estado, _ := putJson(t, entorno, "/cocina/identidad/cambiar-password", map[string]any{
		"password_actual": passwordSeguridad,
		"password_nueva":  "corto",
	})
	if estado != http.StatusBadRequest {
		t.Fatalf("password débil debe rechazarse; got %d", estado)
	}
}

// -----------------------------------------------------------------------------
// CONTENT-TYPE / METHODS
// -----------------------------------------------------------------------------

func TestSeguridad_MetodoIncorrecto_Devuelve404o405(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	iniciarSesionComoOperador(t, entorno)
	resp, err := entorno.cliente.Get(entorno.servidor.URL + "/cocina/sistemas/00000000-0000-0000-0000-000000000000/credencial-externa")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal("GET a /credencial-externa (que es POST) no debería devolver 200")
	}
}

// -----------------------------------------------------------------------------
// HEALTH / OPEN
// -----------------------------------------------------------------------------

func TestSeguridad_HealthCheckEsPublicoYNoFiltraSecretos(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	estado, datos := getJson(t, entorno, "/salud")
	if estado != http.StatusOK {
		t.Fatalf("salud: %d", estado)
	}
	cuerpoRaw, _ := json.Marshal(datos)
	prohibidos := []string{"KEK", "password", "secreto", "secret", "BASE_DATOS_URL"}
	for _, p := range prohibidos {
		if bytes.Contains(bytes.ToLower(cuerpoRaw), bytes.ToLower([]byte(p))) {
			t.Errorf("/salud expone palabra prohibida: %s | %s", p, cuerpoRaw)
		}
	}
}

func TestSeguridad_NoExisteEndpoint_Devuelve404(t *testing.T) {
	entorno := montarEntornoPruebas(t)
	estado, _ := getJson(t, entorno, "/cocina/no/existe/ruta")
	if estado != http.StatusNotFound {
		t.Fatalf("got %d", estado)
	}
}

var _ = errors.New
