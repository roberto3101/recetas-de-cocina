package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sistemas-unificados/capacidades/consumo_usuarios"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoProvisionarUsuario struct {
	CorreoElectronico string `json:"correo_electronico"`
	PasswordPlana     string `json:"password_plana"`
	NombreCompleto    string `json:"nombre_completo"`
}

func ConstruirHandlerProvisionarUsuarioExterno(conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
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
		var cuerpo cuerpoProvisionarUsuario
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}
		usuario, err := consumo_usuarios.ProvisionarUsuarioEnSistema(peticion.Context(), conexion, resolver, clavesCifrado, consumo_usuarios.DatosProvisionarUsuarioExterno{
			SistemaId:         sistemaId,
			CorreoElectronico: cuerpo.CorreoElectronico,
			PasswordPlana:     cuerpo.PasswordPlana,
			NombreCompleto:    cuerpo.NombreCompleto,
			OperadorId:        sesion.UsuarioId,
			SesionId:          &sesion.SesionId,
			IpOrigen:          obtenerIpRemota(peticion),
			AgenteUsuario:     peticion.UserAgent(),
		})
		if err != nil {
			ResponderError(escritor, http.StatusBadGateway, errores.CodigoConexionExternaFallida, err.Error())
			return
		}
		ResponderExito(escritor, http.StatusCreated, usuario)
		_ = adaptadores.ClaveCodeplexVentas // referencia para mantener import si llegara a salir no usado
	}
}
