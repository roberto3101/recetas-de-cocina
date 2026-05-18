package cripto

import (
	"crypto/hmac"
	"crypto/sha256"
	"strings"
)

// HmacSha256 produce un MAC determinístico de `mensaje` con `clave`. Se usa
// para tener un valor determinístico (= mismo input siempre mismo output) que
// puede indexarse en BD, mientras el valor cifrado en sí está protegido con
// AES-GCM (no determinístico). Útil para UNIQUE / lookups por campos sensibles.
//
// Convención: el mensaje se normaliza con TrimSpace + lower antes del HMAC para
// que (correo) <"Foo@bar.com"> y <"foo@bar.com "> matcheen al mismo hash.
func HmacSha256(clave []byte, mensaje string) []byte {
	h := hmac.New(sha256.New, clave)
	h.Write([]byte(strings.ToLower(strings.TrimSpace(mensaje))))
	return h.Sum(nil)
}
