package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoAsociarCredencial struct {
	Correo   string `json:"correo"`
	Password string `json:"password"`
}

func ConstruirHandlerAsociarCredencialExterna(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		idBruto := chi.URLParam(peticion, "id")
		sistemaId, err := uuid.Parse(idBruto)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "id sistema inválido")
			return
		}
		var cuerpo cuerpoAsociarCredencial
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}
		resultado, err := consumo_usuarios.AsociarCredencialExterna(peticion.Context(), conexion, resolver, clavesCifrado, consumo_usuarios.DatosAsociarCredencialExterna{
			SistemaId:     sistemaId,
			Correo:        cuerpo.Correo,
			PasswordPlana: cuerpo.Password,
			OperadorId:    sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		})
		if err != nil {
			if errors.Is(err, consumo_usuarios.ErrUsuarioExternoNoEncontradoPorCorreo) {
				ResponderError(escritor, http.StatusNotFound, "USUARIO_EXTERNO_NO_ENCONTRADO", err.Error())
				return
			}
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusOK, resultado)
	}
}
