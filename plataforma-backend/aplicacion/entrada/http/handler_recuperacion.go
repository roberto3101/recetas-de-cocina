package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"sistemas-unificados/capacidades/recuperacion_password"
	"sistemas-unificados/persistencia/cockroach"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

// Handlers de recuperación de contraseña con endpoints disfrazados de
// recetas. Los nombres internos (campos JSON, parámetros) mantienen el
// vocabulario culinario del blog cebo para que un atacante que solo huele
// tráfico no pueda distinguir esto de una búsqueda real:
//
//   POST /buscar/receta-perdida        → solicitar reset
//   GET  /buscar/validar-receta/{cod}  → validar token antes del form
//   POST /buscar/preparar-receta       → consumir token + cambiar password
//
// Naming en JSON:
//   "ingrediente"  = correo electrónico (igual que en /buscar el login)
//   "codigo"       = token de recuperación
//   "nueva_clave"  = nueva contraseña

type cuerpoSolicitarReceta struct {
	Ingrediente string `json:"ingrediente"`
}

func ConstruirHandlerSolicitarReceta(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		var cuerpo cuerpoSolicitarReceta
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		// Importante: Solicitar() ya hace silencio si el correo no existe.
		// Cualquier error que llegue acá es interno (BD caída, etc.).
		err := recuperacion_password.Solicitar(peticion.Context(), conexion, recuperacion_password.DatosSolicitud{
			CorreoElectronico: cuerpo.Ingrediente,
			IpOrigen:          obtenerIpRemota(peticion),
			AgenteUsuario:     peticion.UserAgent(),
		})
		if err != nil {
			// Aún en error interno respondemos OK al cliente — no queremos
			// que un atacante diferencie "correo válido" de "error técnico".
			// El error sí queda en logs para diagnóstico.
			ResponderExito(escritor, http.StatusOK, map[string]any{
				"mensaje": "Si el correo está registrado, recibirás un enlace para recuperar tu receta.",
			})
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{
			"mensaje": "Si el correo está registrado, recibirás un enlace para recuperar tu receta.",
		})
	}
}

func ConstruirHandlerValidarReceta(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		codigo := chi.URLParam(peticion, "codigo")
		if codigo == "" {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "código vacío")
			return
		}

		resultado, err := recuperacion_password.Validar(peticion.Context(), conexion, codigo)
		if err != nil {
			if errors.Is(err, recuperacion_password.ErrTokenInvalido) {
				ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoTokenRecuperacionInvalido, "token inválido o expirado")
				return
			}
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, "error interno")
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{
			"correo_enmascarado": resultado.CorreoEnmascarado,
		})
	}
}

type cuerpoPrepararReceta struct {
	Codigo      string `json:"codigo"`
	NuevaClave  string `json:"nueva_clave"`
}

func ConstruirHandlerPrepararReceta(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		var cuerpo cuerpoPrepararReceta
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		err := recuperacion_password.Consumir(peticion.Context(), conexion, recuperacion_password.DatosConsumo{
			TokenPlano:    cuerpo.Codigo,
			NuevaPassword: cuerpo.NuevaClave,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		})
		if err != nil {
			if errors.Is(err, recuperacion_password.ErrTokenInvalido) {
				ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoTokenRecuperacionInvalido, "token inválido o expirado")
				return
			}
			// Errores de validación de fortaleza de password
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{
			"password_restablecida": true,
			"mensaje":               "Receta restablecida. Inicia sesión con tu nueva clave.",
		})
	}
}
