package http

import (
	"time"

	"github.com/go-chi/chi/v5"

	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

type DependenciasRutas struct {
	ConexionBaseDatos *cockroach.ConexionBaseDatos
	ClavesCifrado     *cripto.ClavesCifrado
	ResolverAdaptador adaptadores.Resolver
}

func RegistrarRutas(enrutador chi.Router, deps DependenciasRutas) {
	limitadorCebo := NuevoLimitadorVelocidad(60, time.Minute).AplicarComoMiddleware()

	// Salud — fuera del cebo, útil para monitoreo interno
	enrutador.Get("/salud", ConstruirManejadorSalud(deps.ConexionBaseDatos))

	// Blog cebo público
	enrutador.Get("/", ConstruirManejadorBlogCeboInicio())
	enrutador.With(limitadorCebo).Post("/buscar", ConstruirManejadorBlogCeboBuscar(deps.ConexionBaseDatos))

	// Zona privada del gestor
	enrutador.Route("/cocina", func(privado chi.Router) {
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

			// CATALOGO_SISTEMAS
			con2FA.Get("/sistemas", ConstruirHandlerListarSistemas(deps.ConexionBaseDatos))
			con2FA.Post("/sistemas", ConstruirHandlerRegistrarSistema(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Get("/sistemas/{id}", ConstruirHandlerConsultarSistema(deps.ConexionBaseDatos))
			con2FA.Put("/sistemas/{id}", ConstruirHandlerActualizarSistema(deps.ConexionBaseDatos))
			con2FA.Delete("/sistemas/{id}", ConstruirHandlerEliminarSistema(deps.ConexionBaseDatos))
			con2FA.Post("/sistemas/{id}/probar-conexion", ConstruirHandlerProbarConexionSistema(deps.ConexionBaseDatos, deps.ClavesCifrado))

			// CONSUMO_USUARIOS
			con2FA.Get("/usuarios-externos", ConstruirHandlerListarUsuariosAgregados(deps.ConexionBaseDatos, deps.ResolverAdaptador))
			con2FA.Get("/diagnostico/sistemas-externos", ConstruirHandlerDiagnosticoExternos(deps.ConexionBaseDatos, deps.ResolverAdaptador))
			con2FA.Get("/sistemas/{id}/usuarios", ConstruirHandlerListarUsuariosDeSistema(deps.ConexionBaseDatos, deps.ResolverAdaptador))
			con2FA.Get("/sistemas/{id}/usuarios/{usuario_id}", ConstruirHandlerConsultarUsuarioExterno(deps.ConexionBaseDatos, deps.ResolverAdaptador))
			con2FA.Post("/sistemas/{id}/usuarios", ConstruirHandlerProvisionarUsuarioExterno(deps.ConexionBaseDatos, deps.ResolverAdaptador, deps.ClavesCifrado))
			con2FA.Post("/sistemas/{id}/credencial-externa", ConstruirHandlerAsociarCredencialExterna(deps.ConexionBaseDatos, deps.ResolverAdaptador, deps.ClavesCifrado))
			con2FA.Put("/sistemas/{id}/usuarios/{usuario_id}", ConstruirHandlerEditarUsuarioExterno(deps.ConexionBaseDatos, deps.ResolverAdaptador))
			con2FA.Put("/sistemas/{id}/usuarios/{usuario_id}/password", ConstruirHandlerCambiarPasswordUsuarioExterno(deps.ConexionBaseDatos, deps.ResolverAdaptador, deps.ClavesCifrado))
			con2FA.Get("/sistemas/{id}/usuarios/{usuario_id}/autofill", ConstruirHandlerAutofill(deps.ConexionBaseDatos, deps.ClavesCifrado))
			con2FA.Get("/sistemas/{id}/usuarios/{usuario_id}/autofill-hash", ConstruirHandlerAutofillConHash(deps.ConexionBaseDatos, deps.ResolverAdaptador))

			// AUDITORIA
			con2FA.Get("/auditoria", ConstruirHandlerConsultarAuditoria(deps.ConexionBaseDatos))
		})
	})
}
