package http

import (
	"context"
	"net/http"
	"time"

	"sistemas-unificados/persistencia/cockroach"
)

type respuestaSalud struct {
	Servicio   string `json:"servicio"`
	Version    string `json:"version"`
	BaseDatos  string `json:"base_datos"`
	MarcaHora  string `json:"marca_hora"`
}

func ConstruirManejadorSalud(conexionBaseDatos *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		contextoSalud, cancelar := context.WithTimeout(peticion.Context(), 2*time.Second)
		defer cancelar()

		estadoBaseDatos := "ok"
		if err := conexionBaseDatos.VerificarSalud(contextoSalud); err != nil {
			estadoBaseDatos = "indisponible: " + err.Error()
			ResponderError(escritor, http.StatusServiceUnavailable, "BASE_DATOS_INDISPONIBLE", estadoBaseDatos)
			return
		}

		ResponderExito(escritor, http.StatusOK, respuestaSalud{
			Servicio:  "plataforma-backend",
			Version:   "0.1.0",
			BaseDatos: estadoBaseDatos,
			MarcaHora: time.Now().UTC().Format(time.RFC3339),
		})
	}
}
