package cripto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerarTokenSesionAleatorio() (tokenPlano string, tokenHash string, err error) {
	bytesAleatorios := make([]byte, 32)
	if _, err := rand.Read(bytesAleatorios); err != nil {
		return "", "", err
	}

	tokenPlano = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytesAleatorios)
	tokenHash = HashearTokenConSha256(tokenPlano)
	return tokenPlano, tokenHash, nil
}

func HashearTokenConSha256(tokenPlano string) string {
	resumen := sha256.Sum256([]byte(tokenPlano))
	return hex.EncodeToString(resumen[:])
}
