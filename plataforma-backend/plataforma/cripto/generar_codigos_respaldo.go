package cripto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const cantidadCodigosRespaldo = 10
const longitudBytesCodigoRespaldo = 5

func GenerarCodigosRespaldoTotp() ([]string, []string, error) {
	codigosPlanos := make([]string, 0, cantidadCodigosRespaldo)
	codigosHash := make([]string, 0, cantidadCodigosRespaldo)

	for i := 0; i < cantidadCodigosRespaldo; i++ {
		bytes := make([]byte, longitudBytesCodigoRespaldo)
		if _, err := rand.Read(bytes); err != nil {
			return nil, nil, err
		}
		codigoPlano := fmt.Sprintf("%s-%s", hex.EncodeToString(bytes[:2]), hex.EncodeToString(bytes[2:]))
		codigosPlanos = append(codigosPlanos, codigoPlano)
		codigosHash = append(codigosHash, HashearTokenConSha256(codigoPlano))
	}

	return codigosPlanos, codigosHash, nil
}
