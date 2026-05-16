package identidad

import (
	"time"

	"github.com/google/uuid"
)

const (
	EstadoSesionActiva                 = "ACTIVA"
	EstadoSesionExpirada               = "EXPIRADA"
	EstadoSesionRevocada               = "REVOCADA"
	EstadoSesionInvalidada             = "INVALIDADA"
	EstadoSesionPendienteSegundoFactor = "PENDIENTE_SEGUNDO_FACTOR"
)

type SesionGlobal struct {
	Id                    uuid.UUID
	UsuarioId             uuid.UUID
	TokenHash             string
	RefreshTokenHash      string
	DispositivoId         string
	IpOrigen              string
	AgenteUsuario         string
	SegundoFactorValidado bool
	EmitidaEn             time.Time
	ExpiraEn              time.Time
	Estado                string
	UltimoAccesoEn        *time.Time
	RevocadoEn            *time.Time
	RevocadoPor           *uuid.UUID
	MotivoRevocacion      string
}
