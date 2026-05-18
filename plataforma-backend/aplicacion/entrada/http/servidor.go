package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

type Servidor struct {
	servidorHttp *http.Server
}

type OpcionesServidor struct {
	Direccion         string
	ConexionBaseDatos *cockroach.ConexionBaseDatos
	ClavesCifrado     *cripto.ClavesCifrado
}

func ConstruirServidor(opciones OpcionesServidor) *Servidor {
	enrutador := chi.NewRouter()
	enrutador.Use(middleware.Recoverer)
	enrutador.Use(middleware.RequestID)
	enrutador.Use(RegistrarPeticionEntrante)
	enrutador.Use(middleware.Timeout(30 * time.Second))

	RegistrarRutas(enrutador, DependenciasRutas{
		ConexionBaseDatos: opciones.ConexionBaseDatos,
		ClavesCifrado:     opciones.ClavesCifrado,
	})

	servidorHttp := &http.Server{
		Addr:              opciones.Direccion,
		Handler:           enrutador,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return &Servidor{servidorHttp: servidorHttp}
}

func (s *Servidor) Direccion() string {
	return s.servidorHttp.Addr
}

func (s *Servidor) EscucharHasta(contexto context.Context) error {
	errEscucha := make(chan error, 1)
	go func() {
		if err := s.servidorHttp.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errEscucha <- err
		}
		close(errEscucha)
	}()

	select {
	case <-contexto.Done():
		contextoApagado, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancelar()
		return s.servidorHttp.Shutdown(contextoApagado)
	case err := <-errEscucha:
		return err
	}
}
