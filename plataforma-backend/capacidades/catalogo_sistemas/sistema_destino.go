package catalogo_sistemas

import (
	"time"

	"github.com/google/uuid"
)

const (
	EstadoSistemaActivo    = "ACTIVO"
	EstadoSistemaInactivo  = "INACTIVO"
	EstadoSistemaEliminado = "ELIMINADO"

	MetodoLoginPost = "POST"
	MetodoLoginGet  = "GET"
)

type SistemaDestino struct {
	Id                  uuid.UUID
	Codigo              string
	Nombre              string
	UrlAcceso           string
	UrlLogin            string
	NombreCampoUsuario  string
	NombreCampoPassword string
	MetodoLogin         string
	Estado              string
	CreadoEn            time.Time
	CreadoPor           *uuid.UUID
	ActualizadoEn       *time.Time
	ActualizadoPor      *uuid.UUID
}
