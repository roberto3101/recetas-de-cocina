package http

import (
	"net/http"
)

type respuestaPerfil struct {
	UsuarioId             string `json:"usuario_id"`
	CorreoElectronico     string `json:"correo_electronico"`
	SegundoFactorValidado bool   `json:"segundo_factor_validado"`
	ExpiraEn              string `json:"expira_en"`
}

func ConstruirHandlerConsultarPerfil() http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}
		ResponderExito(escritor, http.StatusOK, respuestaPerfil{
			UsuarioId:             sesion.UsuarioId.String(),
			CorreoElectronico:     sesion.CorreoElectronico,
			SegundoFactorValidado: sesion.SegundoFactorValidado,
			ExpiraEn:              sesion.ExpiraEn.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
}
