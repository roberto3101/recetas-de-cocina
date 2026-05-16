package validaciones

import (
	"errors"
	"unicode"
)

var (
	ErrPasswordMuyCorto       = errors.New("la contraseña debe tener al menos 12 caracteres")
	ErrPasswordSinMayuscula   = errors.New("la contraseña debe contener al menos una letra mayúscula")
	ErrPasswordSinMinuscula   = errors.New("la contraseña debe contener al menos una letra minúscula")
	ErrPasswordSinDigito      = errors.New("la contraseña debe contener al menos un dígito")
	ErrPasswordSinEspecial    = errors.New("la contraseña debe contener al menos un carácter especial")
)

func ValidarFortalezaPassword(passwordPlana string) error {
	if len(passwordPlana) < 12 {
		return ErrPasswordMuyCorto
	}
	var tieneMayuscula, tieneMinuscula, tieneDigito, tieneEspecial bool
	for _, r := range passwordPlana {
		switch {
		case unicode.IsUpper(r):
			tieneMayuscula = true
		case unicode.IsLower(r):
			tieneMinuscula = true
		case unicode.IsDigit(r):
			tieneDigito = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			tieneEspecial = true
		}
	}
	if !tieneMayuscula {
		return ErrPasswordSinMayuscula
	}
	if !tieneMinuscula {
		return ErrPasswordSinMinuscula
	}
	if !tieneDigito {
		return ErrPasswordSinDigito
	}
	if !tieneEspecial {
		return ErrPasswordSinEspecial
	}
	return nil
}
