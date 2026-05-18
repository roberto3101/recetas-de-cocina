package http

import (
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

// ConstruirHandlerAutofillAcceso devuelve una página HTML con un form oculto
// que se auto-envía al login del sistema externo con las credenciales descifradas.
func ConstruirHandlerAutofillAcceso(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id inválido")
			return
		}
		acceso, err := boveda.ConsultarAccesoPorId(peticion.Context(), conexion.Pool(), clavesCifrado, id)
		if err != nil {
			if errors.Is(err, boveda.ErrAccesoNoEncontrado) {
				ResponderError(escritor, http.StatusNotFound, "ACCESO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(peticion.Context(), conexion.Pool(), acceso.SistemaDestinoId)
		if err != nil {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoSistemaNoEncontrado, err.Error())
			return
		}
		urlLogin := sistema.UrlLogin
		if urlLogin == "" {
			urlLogin = sistema.UrlAcceso
		}

		trans, err := conexion.Pool().Begin(peticion.Context())
		if err == nil {
			_ = auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
				UsuarioId: &sesion.UsuarioId,
				SesionId:  &sesion.SesionId,
				Modulo:    "AUTOFILL",
				Accion:    "AUTOFILL_GENERADO",
				Entidad:   "acceso_guardado",
				EntidadId: &acceso.Id,
				DatosNuevos: map[string]any{
					"sistema_codigo": sistema.Codigo,
				},
				IpOrigen:      obtenerIpRemota(peticion),
				AgenteUsuario: peticion.UserAgent(),
			})
			_ = trans.Commit(peticion.Context())
		}

		paginaHtml := construirPaginaAutofillSimple(
			urlLogin,
			sistema.MetodoLogin,
			sistema.NombreCampoUsuario,
			sistema.NombreCampoPassword,
			acceso.UsuarioExterno,
			acceso.PasswordPlana,
			sistema.Nombre,
		)
		escritor.Header().Set("Content-Type", "text/html; charset=utf-8")
		escritor.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		escritor.WriteHeader(http.StatusOK)
		_, _ = escritor.Write([]byte(paginaHtml))
	}
}

// ConstruirHandlerAutofillBookmarklet devuelve JSON con usuario y password en
// claro, para que un bookmarklet pueda rellenar el form del sistema externo
// desde la misma pestaña del usuario.
func ConstruirHandlerAutofillBookmarklet(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id inválido")
			return
		}
		acceso, err := boveda.ConsultarAccesoPorId(peticion.Context(), conexion.Pool(), clavesCifrado, id)
		if err != nil {
			if errors.Is(err, boveda.ErrAccesoNoEncontrado) {
				ResponderError(escritor, http.StatusNotFound, "ACCESO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(peticion.Context(), conexion.Pool(), acceso.SistemaDestinoId)
		if err != nil {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoSistemaNoEncontrado, err.Error())
			return
		}

		trans, err := conexion.Pool().Begin(peticion.Context())
		if err == nil {
			_ = auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
				UsuarioId: &sesion.UsuarioId,
				SesionId:  &sesion.SesionId,
				Modulo:    "AUTOFILL",
				Accion:    "BOOKMARKLET_LEYO_CREDENCIAL",
				Entidad:   "acceso_guardado",
				EntidadId: &acceso.Id,
				DatosNuevos: map[string]any{
					"sistema_codigo": sistema.Codigo,
				},
				IpOrigen:      obtenerIpRemota(peticion),
				AgenteUsuario: peticion.UserAgent(),
			})
			_ = trans.Commit(peticion.Context())
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{
			"usuario":               acceso.UsuarioExterno,
			"password":              acceso.PasswordPlana,
			"nombre_campo_usuario":  sistema.NombreCampoUsuario,
			"nombre_campo_password": sistema.NombreCampoPassword,
			"url_login":             sistema.UrlLogin,
			"url_acceso":            sistema.UrlAcceso,
		})
	}
}

func construirPaginaAutofillSimple(urlLogin, metodo, campoUsuario, campoPassword, valorUsuario, valorPassword, nombreSistema string) string {
	if metodo == "" {
		metodo = "POST"
	}
	if campoUsuario == "" {
		campoUsuario = "usuario"
	}
	if campoPassword == "" {
		campoPassword = "password"
	}
	return fmt.Sprintf(`<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>Abriendo %[1]s…</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 480px; margin: 80px auto; padding: 0 16px; color: #333; text-align: center; }
  .spinner { display: inline-block; width: 36px; height: 36px; border: 3px solid #ddd; border-top-color: #8b3a2a; border-radius: 50%%; animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  h1 { color: #8b3a2a; }
</style>
</head>
<body>
  <div class="spinner"></div>
  <h1>Abriendo %[1]s</h1>
  <p>Iniciando sesión automáticamente…</p>

  <form id="formAutofill" method="%[2]s" action="%[3]s" style="display:none">
    <input type="text" name="%[4]s" value="%[5]s">
    <input type="password" name="%[6]s" value="%[7]s">
  </form>

  <script>setTimeout(function () { document.getElementById("formAutofill").submit(); }, 80);</script>
</body>
</html>`,
		html.EscapeString(nombreSistema),
		html.EscapeString(metodo),
		html.EscapeString(urlLogin),
		html.EscapeString(campoUsuario),
		html.EscapeString(valorUsuario),
		html.EscapeString(campoPassword),
		html.EscapeString(valorPassword),
	)
}
