package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoEditarUsuarioExterno struct {
	CorreoElectronico string `json:"correo_electronico"`
	NombreCompleto    string `json:"nombre_completo"`
}

func ConstruirHandlerEditarUsuarioExterno(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		sistemaId, err := uuid.Parse(chi.URLParam(peticion, "id"))
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id sistema inválido")
			return
		}
		idExterno := chi.URLParam(peticion, "usuario_id")
		if idExterno == "" {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "usuario_id requerido")
			return
		}

		var cuerpo cuerpoEditarUsuarioExterno
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		usuario, err := consumo_usuarios.EditarUsuarioEnSistema(peticion.Context(), conexion, resolver, consumo_usuarios.DatosEditarUsuarioExterno{
			SistemaId:              sistemaId,
			IdExterno:              idExterno,
			CorreoElectronicoNuevo: cuerpo.CorreoElectronico,
			NombreCompletoNuevo:    cuerpo.NombreCompleto,
			OperadorId:             sesion.UsuarioId,
			SesionId:               &sesion.SesionId,
			IpOrigen:               obtenerIpRemota(peticion),
			AgenteUsuario:          peticion.UserAgent(),
		})
		if err != nil {
			ResponderError(escritor, http.StatusBadGateway, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, usuario)
	}
}
