package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type CuerpoExito struct {
	Exito bool `json:"exito"`
	Datos any  `json:"datos,omitempty"`
}

type CuerpoError struct {
	Exito       bool   `json:"exito"`
	Error       string `json:"error"`
	CodigoError string `json:"codigo_error,omitempty"`
}

func ResponderExito(escritor http.ResponseWriter, codigoHttp int, datos any) {
	escritor.Header().Set("Content-Type", "application/json; charset=utf-8")
	escritor.WriteHeader(codigoHttp)
	if err := json.NewEncoder(escritor).Encode(CuerpoExito{Exito: true, Datos: datos}); err != nil {
		slog.Error("no se pudo serializar respuesta de éxito", "detalle", err.Error())
	}
}

func ResponderError(escritor http.ResponseWriter, codigoHttp int, codigoError, mensaje string) {
	escritor.Header().Set("Content-Type", "application/json; charset=utf-8")
	escritor.WriteHeader(codigoHttp)
	if err := json.NewEncoder(escritor).Encode(CuerpoError{Exito: false, Error: mensaje, CodigoError: codigoError}); err != nil {
		slog.Error("no se pudo serializar respuesta de error", "detalle", err.Error())
	}
}

func ResponderRecetaInocua(escritor http.ResponseWriter, codigoHttp int, cuerpoHtml string) {
	escritor.Header().Set("Content-Type", "text/html; charset=utf-8")
	escritor.WriteHeader(codigoHttp)
	if _, err := escritor.Write([]byte(cuerpoHtml)); err != nil {
		slog.Error("no se pudo escribir respuesta del blog cebo", "detalle", err.Error())
	}
}
