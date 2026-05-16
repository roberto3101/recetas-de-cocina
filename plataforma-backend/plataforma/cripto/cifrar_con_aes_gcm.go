package cripto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func CifrarConAesGcm(clave []byte, textoPlano []byte) ([]byte, error) {
	bloque, err := aes.NewCipher(clave)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	cifradoConEtiqueta := gcm.Seal(nil, nonce, textoPlano, nil)

	resultado := make([]byte, 0, len(nonce)+len(cifradoConEtiqueta))
	resultado = append(resultado, nonce...)
	resultado = append(resultado, cifradoConEtiqueta...)

	return resultado, nil
}

func DescifrarConAesGcm(clave []byte, paqueteCifrado []byte) ([]byte, error) {
	bloque, err := aes.NewCipher(clave)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return nil, err
	}

	tamanoNonce := gcm.NonceSize()
	if len(paqueteCifrado) < tamanoNonce {
		return nil, errors.New("paquete cifrado demasiado corto: no contiene nonce")
	}

	nonce := paqueteCifrado[:tamanoNonce]
	cifradoConEtiqueta := paqueteCifrado[tamanoNonce:]

	textoPlano, err := gcm.Open(nil, nonce, cifradoConEtiqueta, nil)
	if err != nil {
		return nil, errors.New("descifrado falló: dato manipulado o clave incorrecta")
	}

	return textoPlano, nil
}
