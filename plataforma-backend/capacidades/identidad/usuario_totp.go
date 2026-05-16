package identidad

import (
	"time"

	"github.com/google/uuid"
)

const (
	EstadoTotpPendiente = "PENDIENTE"
	EstadoTotpActivo    = "ACTIVO"
	EstadoTotpRevocado  = "REVOCADO"
)

type UsuarioTotp struct {
	Id              uuid.UUID
	UsuarioId       uuid.UUID
	SecretoCifrado  []byte
	CodigosRespaldo []string
	Estado          string
	ActivadoEn      *time.Time
	RevocadoEn      *time.Time
	CreadoEn        time.Time
}
