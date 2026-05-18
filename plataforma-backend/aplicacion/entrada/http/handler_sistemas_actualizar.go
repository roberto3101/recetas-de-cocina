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
	UrlAcceso           string `json:"url_acceso"`
	UrlLogin            string `json:"url_login"`
	NombreCampoUsuario  string `json:"nombre_campo_usuario"`
	NombreCampoPassword string `json:"nombre_campo_password"`
	MetodoLogin         string `json:"metodo_login"`
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
			UrlAcceso:           cuerpo.UrlAcceso,
			UrlLogin:            cuerpo.UrlLogin,
			NombreCampoUsuario:  cuerpo.NombreCampoUsuario,
			NombreCampoPassword: cuerpo.NombreCampoPassword,
			MetodoLogin:         cuerpo.MetodoLogin,
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
