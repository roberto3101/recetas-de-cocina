package sesion

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SesionActual struct {
	SesionId                 uuid.UUID
	UsuarioId                uuid.UUID
	CorreoElectronico        string
	EmitidaEn                time.Time
	ExpiraEn                 time.Time
	SegundoFactorValidado    bool
	IpOrigen                 string
	AgenteUsuario            string
}

type llaveContextoSesion struct{}

func ContextoConSesion(contexto context.Context, sesion *SesionActual) context.Context {
	return context.WithValue(contexto, llaveContextoSesion{}, sesion)
}

func SesionDesdeContexto(contexto context.Context) (*SesionActual, error) {
	valor := contexto.Value(llaveContextoSesion{})
	if valor == nil {
		return nil, errors.New("no hay sesión activa en el contexto")
	}
	sesion, ok := valor.(*SesionActual)
	if !ok || sesion == nil {
		return nil, errors.New("sesión en contexto con tipo inválido")
	}
	return sesion, nil
}
