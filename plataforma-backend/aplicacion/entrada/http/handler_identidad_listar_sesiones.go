package http

import (
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/errores"
)

type itemSesionListado struct {
	SesionId             string `json:"sesion_id"`
	DispositivoId        string `json:"dispositivo_id"`
	IpOrigen             string `json:"ip_origen"`
	AgenteUsuario        string `json:"agente_usuario"`
	EmitidaEn            string `json:"emitida_en"`
	ExpiraEn             string `json:"expira_en"`
	UltimoAccesoEn       string `json:"ultimo_acceso_en,omitempty"`
	SegundoFactorValidado bool  `json:"segundo_factor_validado"`
}

func ConstruirHandlerListarSesiones(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		activas, err := identidad.ListarSesionesActivasDeUsuario(peticion.Context(), conexion, sesion.UsuarioId)
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, errores.CodigoErrorInterno, err.Error())
			return
		}

		listado := make([]itemSesionListado, 0, len(activas))
		for _, s := range activas {
			ultimoAcceso := ""
			if s.UltimoAccesoEn != nil {
				ultimoAcceso = s.UltimoAccesoEn.UTC().Format("2006-01-02T15:04:05Z")
			}
			listado = append(listado, itemSesionListado{
				SesionId:              s.Id.String(),
				DispositivoId:         s.DispositivoId,
				IpOrigen:              s.IpOrigen,
				AgenteUsuario:         s.AgenteUsuario,
				EmitidaEn:             s.EmitidaEn.UTC().Format("2006-01-02T15:04:05Z"),
				ExpiraEn:              s.ExpiraEn.UTC().Format("2006-01-02T15:04:05Z"),
				UltimoAccesoEn:        ultimoAcceso,
				SegundoFactorValidado: s.SegundoFactorValidado,
			})
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{
			"sesiones": listado,
			"total":    len(listado),
		})
	}
}
