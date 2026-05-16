package sesion

import (
	"net/http"
	"time"
)

const NombreCookieSesion = "sesion_cocina"

func EstablecerCookieSesion(escritor http.ResponseWriter, tokenPlano string, expiracion time.Time) {
	http.SetCookie(escritor, &http.Cookie{
		Name:     NombreCookieSesion,
		Value:    tokenPlano,
		Path:     "/",
		Expires:  expiracion,
		MaxAge:   int(time.Until(expiracion).Seconds()),
		HttpOnly: true,
		Secure:   false, // pasar a true en producción detrás de HTTPS
		SameSite: http.SameSiteStrictMode,
	})
}

func LimpiarCookieSesion(escritor http.ResponseWriter) {
	http.SetCookie(escritor, &http.Cookie{
		Name:     NombreCookieSesion,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}

func LeerTokenDeCookie(peticion *http.Request) string {
	cookie, err := peticion.Cookie(NombreCookieSesion)
	if err != nil {
		return ""
	}
	return cookie.Value
}
