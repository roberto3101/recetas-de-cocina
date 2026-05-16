package cripto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
)

const longitudClaveAES256 = 32

type ClavesCifrado struct {
	claveBoveda []byte
}

func CargarClavesDesdeEntorno() (*ClavesCifrado, error) {
	claveCodificada := os.Getenv("KEK_GESTOR")
	if claveCodificada == "" {
		return nil, errors.New("KEK_GESTOR no está definida en entorno (clave maestra de cifrado de la bóveda)")
	}

	claveBoveda, err := base64.StdEncoding.DecodeString(claveCodificada)
	if err != nil {
		return nil, errors.New("KEK_GESTOR no es base64 válido")
	}

	if len(claveBoveda) != longitudClaveAES256 {
		return nil, errors.New("KEK_GESTOR debe ser una clave AES-256 (32 bytes) codificada en base64")
	}

	return &ClavesCifrado{claveBoveda: claveBoveda}, nil
}

func (c *ClavesCifrado) ClaveBoveda() []byte {
	copia := make([]byte, len(c.claveBoveda))
	copy(copia, c.claveBoveda)
	return copia
}

func GenerarClaveAleatoriaAES256() (string, error) {
	clave := make([]byte, longitudClaveAES256)
	if _, err := rand.Read(clave); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(clave), nil
}
