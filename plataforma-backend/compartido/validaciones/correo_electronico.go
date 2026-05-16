package validaciones

import (
	"net/mail"
	"strings"
)

func EsCorreoElectronicoValido(correo string) bool {
	correoLimpio := strings.TrimSpace(correo)
	if correoLimpio == "" || len(correoLimpio) > 254 {
		return false
	}
	direccion, err := mail.ParseAddress(correoLimpio)
	if err != nil {
		return false
	}
	return direccion.Address == correoLimpio
}

func NormalizarCorreo(correo string) string {
	return strings.ToLower(strings.TrimSpace(correo))
}
