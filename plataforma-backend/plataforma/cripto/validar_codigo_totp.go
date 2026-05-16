package cripto

import (
	"github.com/pquerna/otp/totp"
)

func ValidarCodigoTotp(codigoIngresado, secretoBase32 string) bool {
	return totp.Validate(codigoIngresado, secretoBase32)
}
