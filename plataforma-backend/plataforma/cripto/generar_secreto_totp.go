package cripto

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type ResultadoActivacionTotp struct {
	SecretoBase32       string
	UrlOtpAuth          string
	ImagenCodigoQrPng   []byte
	CodigosRespaldo     []string
	CodigosRespaldoHash []string
}

func GenerarSecretoTotp(emisor, cuenta string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      emisor,
		AccountName: cuenta,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
}
