package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"sistemas-unificados/plataforma/gobierno/errores"
	"sistemas-unificados/plataforma/sesion"
)

const maxBytesCuerpo = 1 << 20

var ErrPeticionMalFormada = errors.New("petición mal formada")

func LeerCuerpoJson(escritor http.ResponseWriter, peticion *http.Request, destino any) bool {
	peticion.Body = http.MaxBytesReader(escritor, peticion.Body, maxBytesCuerpo)
	decodificador := json.NewDecoder(peticion.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		if errors.Is(err, io.EOF) {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "cuerpo vacío")
		} else {
			ResponderError(escritor, http.StatusBadRequest, errores.CodigoPeticionMalFormada, "JSON inválido")
		}
		return false
	}
	return true
}

func ObtenerSesionDelContexto(escritor http.ResponseWriter, peticion *http.Request) (*sesion.SesionActual, bool) {
	s, err := sesion.SesionDesdeContexto(peticion.Context())
	if err != nil {
		ResponderError(escritor, http.StatusUnauthorized, errores.CodigoSesionRequerida, "se requiere sesión activa")
		return nil, false
	}
	return s, true
}

func LeerParametroEnteroConDefault(peticion *http.Request, nombre string, valorDefault, valorMaximo int) int {
	bruto := peticion.URL.Query().Get(nombre)
	if bruto == "" {
		return valorDefault
	}
	entero, err := strconv.Atoi(bruto)
	if err != nil || entero <= 0 {
		return valorDefault
	}
	if valorMaximo > 0 && entero > valorMaximo {
		return valorMaximo
	}
	return entero
}

func LeerParametroUuidOpcional(peticion *http.Request, nombre string) *uuid.UUID {
	bruto := peticion.URL.Query().Get(nombre)
	if bruto == "" {
		return nil
	}
	id, err := uuid.Parse(bruto)
	if err != nil {
		return nil
	}
	return &id
}
