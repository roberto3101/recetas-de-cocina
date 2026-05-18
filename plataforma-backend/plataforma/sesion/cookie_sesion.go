package sesion

import (
	"net/http"
	"os"
	"strings"
	"time"
)

const NombreCookieSesion = "sesion_cocina"

// cookieSecure devuelve true cuando COOKIE_SECURE está en true/1/yes en el entorno.
// En producción (HTTPS) DEBE ir true; en dev local (HTTP en localhost) va false.
func cookieSecure() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE")))
	return v == "true" || v == "1" || v == "yes"
}

func EstablecerCookieSesion(escritor http.ResponseWriter, tokenPlano string, expiracion time.Time) {
	http.SetCookie(escritor, &http.Cookie{
		Name:     NombreCookieSesion,
		Value:    tokenPlano,
		Path:     "/",
		Expires:  expiracion,
		MaxAge:   int(time.Until(expiracion).Seconds()),
		HttpOnly: true,
		Secure:   cookieSecure(),
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
		Secure:   cookieSecure(),
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
