package recuperacion_password

import "errors"

var (
	// ErrTokenInvalido es el error genérico devuelto ante cualquier falla de
	// validación del token (no existe, expiró, ya consumido). Intencionalmente
	// no diferencia el motivo para no filtrar información al atacante.
	ErrTokenInvalido = errors.New("token de recuperación inválido o expirado")
)
