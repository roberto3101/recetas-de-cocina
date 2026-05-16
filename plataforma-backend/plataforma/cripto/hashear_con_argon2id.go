package cripto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	tiempoArgon2id      uint32 = 3
	memoriaArgon2idKiB  uint32 = 64 * 1024
	paralelismoArgon2id uint8  = 4
	longitudHashArgon2  uint32 = 32
	longitudSalArgon2   uint32 = 16
)

func HashearPasswordConArgon2id(passwordPlana string) (string, error) {
	sal := make([]byte, longitudSalArgon2)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(passwordPlana), sal, tiempoArgon2id, memoriaArgon2idKiB, paralelismoArgon2id, longitudHashArgon2)

	salCodificada := base64.RawStdEncoding.EncodeToString(sal)
	hashCodificado := base64.RawStdEncoding.EncodeToString(hash)

	cadenaFormateada := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memoriaArgon2idKiB, tiempoArgon2id, paralelismoArgon2id, salCodificada, hashCodificado)

	return cadenaFormateada, nil
}

func VerificarPasswordContraHashArgon2id(passwordPlana, hashAlmacenado string) (bool, error) {
	partes := strings.Split(hashAlmacenado, "$")
	if len(partes) != 6 || partes[1] != "argon2id" {
		return false, errors.New("formato de hash argon2id inválido")
	}

	var version int
	if _, err := fmt.Sscanf(partes[2], "v=%d", &version); err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("versión de argon2id no soportada")
	}

	var memoria uint32
	var tiempo uint32
	var paralelismo uint8
	if _, err := fmt.Sscanf(partes[3], "m=%d,t=%d,p=%d", &memoria, &tiempo, &paralelismo); err != nil {
		return false, err
	}

	sal, err := base64.RawStdEncoding.DecodeString(partes[4])
	if err != nil {
		return false, err
	}

	hashEsperado, err := base64.RawStdEncoding.DecodeString(partes[5])
	if err != nil {
		return false, err
	}

	hashCalculado := argon2.IDKey([]byte(passwordPlana), sal, tiempo, memoria, paralelismo, uint32(len(hashEsperado)))

	coinciden := subtle.ConstantTimeCompare(hashEsperado, hashCalculado) == 1
	return coinciden, nil
}
