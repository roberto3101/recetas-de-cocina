package recuperacion_password

import (
	"time"

	"github.com/google/uuid"
)

// VigenciaDelToken: ventana corta porque el token vive en email / canal lateral.
// 30 min es estándar para reset-password (suficiente para que el usuario lo
// reciba y use, no tanto como para que valga la pena interceptarlo después).
const VigenciaDelToken = 30 * time.Minute

// LongitudDelToken en bytes aleatorios. 32 bytes = 256 bits = inadivinable.
const LongitudDelToken = 32

type Token struct {
	Id            uuid.UUID
	UsuarioId     uuid.UUID
	TokenHash     string
	IpOrigen      string
	AgenteUsuario string
	CreadoEn      time.Time
	ExpiraEn      time.Time
	ConsumidoEn   *time.Time
}

func (t *Token) EstaVigente(ahora time.Time) bool {
	return t.ConsumidoEn == nil && ahora.Before(t.ExpiraEn)
}
