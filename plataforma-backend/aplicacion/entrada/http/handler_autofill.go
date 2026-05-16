package http

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

// ConstruirHandlerAutofill devuelve una pagina HTML con un form oculto que
// auto-envia las credenciales descifradas al login del sistema externo.
// El operador solo ve "redirigiendo..." durante una fraccion de segundo.
func ConstruirHandlerAutofill(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		sistemaId, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id sistema inválido")
			return
		}
		idExterno := chi.URLParam(peticion, "usuario_id")
		if idExterno == "" {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "usuario_id requerido")
			return
		}

		sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(peticion.Context(), conexion.Pool(), sistemaId)
		if err != nil {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoSistemaNoEncontrado, err.Error())
			return
		}
		if sistema.UrlLogin == "" {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoSistemaInvalido, "el sistema no tiene URL de login configurada")
			return
		}

		credencial, err := catalogo_sistemas.ConsultarCredencialUsuarioExterno(peticion.Context(), conexion.Pool(), sistemaId, idExterno)
		if err != nil {
			if errors.Is(err, catalogo_sistemas.ErrCredencialUsuarioNoEncontrada) {
				ResponderError(escritor, http.StatusNotFound, "CREDENCIAL_NO_GUARDADA",
					"este usuario no tiene contraseña guardada en el gestor. Cambia su contraseña desde aquí para guardarla.")
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		passwordPlana, err := cripto.DescifrarConAesGcm(clavesCifrado.ClaveBoveda(), credencial.PasswordCifrada)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoDescifradoFallido, err.Error())
			return
		}

		correoUsuario := credencial.CorreoUsuario
		if correoUsuario == "" {
			correoUsuario = idExterno
		}

		registrarAutofillEnAuditoria(peticion.Context(), conexion, sistema, sesion.UsuarioId, sesion.SesionId, idExterno, peticion)

		paginaHtml := construirPaginaAutofill(
			sistema.UrlLogin,
			sistema.MetodoLogin,
			sistema.NombreCampoUsuario,
			sistema.NombreCampoPassword,
			correoUsuario,
			string(passwordPlana),
			sistema.Nombre,
		)
		escritor.Header().Set("Content-Type", "text/html; charset=utf-8")
		escritor.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		escritor.WriteHeader(http.StatusOK)
		_, _ = escritor.Write([]byte(paginaHtml))
	}
}

func construirPaginaAutofill(urlLogin, metodo, campoUsuario, campoPassword, valorUsuario, valorPassword, nombreSistema string) string {
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
  .nota { color: #888; font-size: 13px; margin-top: 16px; }
</style>
</head>
<body>
  <div class="spinner"></div>
  <h1>Abriendo %[1]s</h1>
  <p>Iniciando sesión automáticamente…</p>
  <noscript>
    <p style="color:red">JavaScript desactivado. <button form="formAutofill" type="submit">Haz click aquí para entrar</button></p>
  </noscript>
  <p class="nota">Si la página no carga, contacta al administrador del gestor.</p>

  <form id="formAutofill" method="%[2]s" action="%[3]s" style="display:none">
    <input type="text" name="%[4]s" value="%[5]s">
    <input type="password" name="%[6]s" value="%[7]s">
  </form>

  <script>
    setTimeout(function () { document.getElementById("formAutofill").submit(); }, 80);
  </script>
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

func registrarAutofillEnAuditoria(contexto context.Context, conexion *cockroach.ConexionBaseDatos, sistema *catalogo_sistemas.SistemaDestino, operadorId, sesionId uuid.UUID, idExterno string, peticion *http.Request) {
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return
	}
	defer trans.Rollback(contexto)
	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &operadorId,
		SesionId:  &sesionId,
		Modulo:    "AUTOFILL",
		Accion:    "AUTOFILL_GENERADO",
		Entidad:   "sistema_destino",
		EntidadId: &sistema.Id,
		DatosNuevos: map[string]any{
			"sistema_codigo":     sistema.Codigo,
			"id_externo_usuario": idExterno,
		},
		IpOrigen:      obtenerIpRemota(peticion),
		AgenteUsuario: peticion.UserAgent(),
	}); err != nil {
		return
	}
	_ = trans.Commit(contexto)
}
