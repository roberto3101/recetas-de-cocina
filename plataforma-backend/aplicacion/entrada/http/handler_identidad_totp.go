package http

import (
	"errors"
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoCodigoTotp struct {
	Codigo string `json:"codigo"`
}

func ConstruirHandlerIniciarActivacionTotp(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		resultado, err := identidad.IniciarActivacionTotp(peticion.Context(), conexion, clavesCifrado, identidad.DatosIniciarActivacionTotp{
			UsuarioId:         sesion.UsuarioId,
			CorreoElectronico: sesion.CorreoElectronico,
			SesionId:          &sesion.SesionId,
			IpOrigen:          obtenerIpRemota(peticion),
			AgenteUsuario:     peticion.UserAgent(),
		})
		if errors.Is(err, identidad.ErrTotpYaActivo) {
			ResponderError(escritor, http.StatusConflict, codigosError.CodigoTotpYaActivo, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{
			"secreto_base32":    resultado.SecretoBase32,
			"url_otp_auth":      resultado.UrlOtpAuth,
			"codigos_respaldo":  resultado.CodigosRespaldo,
			"totp_id":           resultado.TotpId.String(),
		})
	}
}

func ConstruirHandlerConfirmarActivacionTotp(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		var cuerpo cuerpoCodigoTotp
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		err := identidad.ConfirmarActivacionTotp(peticion.Context(), conexion, clavesCifrado, identidad.DatosConfirmarActivacionTotp{
			UsuarioId:       sesion.UsuarioId,
			CodigoIngresado: cuerpo.Codigo,
			SesionId:        &sesion.SesionId,
			IpOrigen:        obtenerIpRemota(peticion),
			AgenteUsuario:   peticion.UserAgent(),
		})
		if errors.Is(err, identidad.ErrCodigoTotpInvalido) {
			ResponderError(escritor, http.StatusUnauthorized, codigosError.CodigoTotpCodigoInvalido, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"totp_activado": true})
	}
}

func ConstruirHandlerValidarSegundoFactor(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		var cuerpo cuerpoCodigoTotp
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		err := identidad.ValidarSegundoFactor(peticion.Context(), conexion, clavesCifrado, identidad.DatosValidarSegundoFactor{
			UsuarioId:       sesion.UsuarioId,
			SesionId:        sesion.SesionId,
			CodigoIngresado: cuerpo.Codigo,
			IpOrigen:        obtenerIpRemota(peticion),
			AgenteUsuario:   peticion.UserAgent(),
		})
		if errors.Is(err, identidad.ErrCodigoTotpInvalido) {
			ResponderError(escritor, http.StatusUnauthorized, codigosError.CodigoTotpCodigoInvalido, err.Error())
			return
		}
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"segundo_factor_validado": true})
	}
}

func ConstruirHandlerRevocarTotp(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		if err := identidad.RevocarTotpActivo(peticion.Context(), conexion, identidad.DatosRevocarTotp{
			UsuarioId:     sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		}); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		ResponderExito(escritor, http.StatusOK, map[string]any{"totp_revocado": true})
	}
}
