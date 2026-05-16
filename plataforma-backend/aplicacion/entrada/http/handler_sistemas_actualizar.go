package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/persistencia/cockroach"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoActualizarSistema struct {
	Nombre              string `json:"nombre"`
	Descripcion         string `json:"descripcion"`
	UrlAcceso           string `json:"url_acceso"`
	Motor               string `json:"motor"`
	ClaveAdaptador      string `json:"clave_adaptador"`
	RequiereLoginGlobal bool   `json:"requiere_login_global"`
	SoportaLectura      bool   `json:"soporta_lectura"`
	SoportaAutoregistro bool   `json:"soporta_autoregistro"`
	Estado              string `json:"estado"`
}

func ConstruirHandlerActualizarSistema(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		idBruto := chi.URLParam(peticion, "id")
		id, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoPeticionMalFormada, "id inválido")
			return
		}
		var cuerpo cuerpoActualizarSistema
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		err = catalogo_sistemas.ActualizarSistema(peticion.Context(), conexion, catalogo_sistemas.DatosActualizarSistema{
			SistemaId:           id,
			Nombre:              cuerpo.Nombre,
			Descripcion:         cuerpo.Descripcion,
			UrlAcceso:           cuerpo.UrlAcceso,
			Motor:               cuerpo.Motor,
			ClaveAdaptador:      cuerpo.ClaveAdaptador,
			RequiereLoginGlobal: cuerpo.RequiereLoginGlobal,
			SoportaLectura:      cuerpo.SoportaLectura,
			SoportaAutoregistro: cuerpo.SoportaAutoregistro,
			Estado:              cuerpo.Estado,
			ActualizadoPor:      sesion.UsuarioId,
			SesionId:            &sesion.SesionId,
			IpOrigen:            obtenerIpRemota(peticion),
			AgenteUsuario:       peticion.UserAgent(),
		})
		if errors.Is(err, catalogo_sistemas.ErrSistemaNoEncontrado) {
			ResponderError(escritor, http.StatusNotFound, codigosError.CodigoSistemaNoEncontrado, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, map[string]any{"sistema_actualizado": true})
	}
}
