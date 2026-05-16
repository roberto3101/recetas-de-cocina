package identidad

import (
	"time"

	"github.com/google/uuid"
)

const (
	EstadoUsuarioPendiente = "PENDIENTE"
	EstadoUsuarioActivo    = "ACTIVO"
	EstadoUsuarioInactivo  = "INACTIVO"
	EstadoUsuarioBloqueado = "BLOQUEADO"
	EstadoUsuarioEliminado = "ELIMINADO"
)

type Usuario struct {
	Id                          uuid.UUID
	CorreoElectronico           string
	PasswordHash                string
	CorreoElectronicoVerificado bool
	Estado                      string
	IntentosFallidos            int64
	BloqueadoHasta              *time.Time
	UltimoInicioSesionEn        *time.Time
	CreadoEn                    time.Time
	CreadoPor                   *uuid.UUID
	ActualizadoEn               *time.Time
	ActualizadoPor              *uuid.UUID
}
