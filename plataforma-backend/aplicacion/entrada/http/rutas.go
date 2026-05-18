package http

import (
	"time"

	"github.com/go-chi/chi/v5"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

type DependenciasRutas struct {
	ConexionBaseDatos *cockroach.ConexionBaseDatos
	ClavesCifrado     *cripto.ClavesCifrado
}

func RegistrarRutas(enrutador chi.Router, deps DependenciasRutas) {
	limitadorCebo := NuevoLimitadorVelocidad(60, time.Minute).AplicarComoMiddleware()
	// Permisivo para humano (10 req/segundo de promedio), pero corta DoS si una
	// sesión robada intenta hacer miles de requests.
	limitadorAutenticado := NuevoLimitadorVelocidad(600, time.Minute).AplicarComoMiddleware()

	// Salud — fuera del cebo, útil para monitoreo interno
	enrutador.Get("/salud", ConstruirManejadorSalud(deps.ConexionBaseDatos))

	// Blog cebo público
	enrutador.Get("/", ConstruirManejadorBlogCeboInicio())
	enrutador.With(limitadorCebo).Post("/buscar", ConstruirManejadorBlogCeboBuscar(deps.ConexionBaseDatos))

	// Zona privada del gestor
	enrutador.Route("/cocina", func(privado chi.Router) {
		privado.Use(limitadorAutenticado)
		privado.Use(RequiereSesion(deps.ConexionBaseDatos))

		// Páginas placeholder (SPA después)
		privado.Get("/inventario", ConstruirManejadorCocinaInicio())
		privado.Get("/verificar", ConstruirManejadorCocinaVerificar())

		// IDENTIDAD
		privado.Get("/identidad/perfil", ConstruirHandlerConsultarPerfil())
		privado.Post("/identidad/cerrar-sesion", ConstruirHandlerCerrarSesion(deps.ConexionBaseDatos))
		privado.Get("/identidad/sesiones", ConstruirHandlerListarSesiones(deps.ConexionBaseDatos))

		// SEGUNDO FACTOR
		privado.Post("/identidad/totp/iniciar-activacion", ConstruirHandlerIniciarActivacionTotp(deps.ConexionBaseDatos, deps.ClavesCifrado))
		privado.Post("/identidad/totp/confirmar-activacion", ConstruirHandlerConfirmarActivacionTotp(deps.ConexionBaseDatos, deps.ClavesCifrado))
		privado.Post("/identidad/totp/validar", ConstruirHandlerValidarSegundoFactor(deps.ConexionBaseDatos, deps.ClavesCifrado))

		// Operaciones que requieren 2FA validado
		privado.Group(func(con2FA chi.Router) {
			con2FA.Use(RequiereSegundoFactor)

			con2FA.Put("/identidad/cambiar-password", ConstruirHandlerCambiarPassword(deps.ConexionBaseDatos))
			con2FA.Delete("/identidad/sesiones/{id}", ConstruirHandlerRevocarSesion(deps.ConexionBaseDatos))
			con2FA.Post("/identidad/totp/revocar", ConstruirHandlerRevocarTotp(deps.ConexionBaseDatos))
			con2FA.Post("/identidad/registrar-operador", ConstruirHandlerRegistrarOperador(deps.ConexionBaseDatos))

			// CATALOGO_SISTEMAS — URLs registradas + selectores del form de login
			con2FA.Get("/sistemas", ConstruirHandlerListarSistemas(deps.ConexionBaseDatos))
			con2FA.Post("/sistemas", ConstruirHandlerRegistrarSistema(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Get("/sistemas/{id}", ConstruirHandlerConsultarSistema(deps.ConexionBaseDatos))
			con2FA.Put("/sistemas/{id}", ConstruirHandlerActualizarSistema(deps.ConexionBaseDatos))
			con2FA.Delete("/sistemas/{id}", ConstruirHandlerEliminarSistema(deps.ConexionBaseDatos))

			// BOVEDA — accesos guardados (titulo + sistema + usuario + clave cifrada)
			con2FA.Get("/boveda/accesos", ConstruirHandlerListarAccesos(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Post("/boveda/accesos", ConstruirHandlerGuardarAcceso(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Put("/boveda/accesos/{id}", ConstruirHandlerEditarAcceso(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Delete("/boveda/accesos/{id}", ConstruirHandlerDesactivarAcceso(deps.ConexionBaseDatos))
			con2FA.Post("/boveda/accesos/{id}/reactivar", ConstruirHandlerReactivarAcceso(deps.ConexionBaseDatos))
			con2FA.Get("/boveda/accesos/{id}/autofill", ConstruirHandlerAutofillAcceso(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Get("/boveda/accesos/{id}/bookmarklet", ConstruirHandlerAutofillBookmarklet(deps.ConexionBaseDatos, deps.ClavesCifrado))

			// AUDITORIA
			con2FA.Get("/auditoria", ConstruirHandlerConsultarAuditoria(deps.ConexionBaseDatos))
		})
	})
}
