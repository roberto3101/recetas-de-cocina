package catalogo_sistemas

import (
	"time"

	"github.com/google/uuid"
)

const (
	MotorCockroachdb = "COCKROACHDB"
	MotorPostgresql  = "POSTGRESQL"
	MotorMysql       = "MYSQL"
	MotorMariadb     = "MARIADB"
	MotorApiRest     = "API_REST"
	MotorOtro        = "OTRO"

	EstadoSistemaActivo    = "ACTIVO"
	EstadoSistemaInactivo  = "INACTIVO"
	EstadoSistemaEliminado = "ELIMINADO"

	AlgoritmoBcrypt    = "BCRYPT"
	AlgoritmoArgon2id  = "ARGON2ID"
	AlgoritmoSha256Hex = "SHA256_HEX"
	AlgoritmoSha512Hex = "SHA512_HEX"
	AlgoritmoPbkdf2    = "PBKDF2"
	AlgoritmoPlaintext = "PLAINTEXT"
	AlgoritmoOtro      = "OTRO"
)

type SistemaDestino struct {
	Id                  uuid.UUID
	Codigo              string
	Nombre              string
	Descripcion         string
	UrlAcceso           string
	UrlLogin            string
	NombreCampoUsuario  string
	NombreCampoPassword string
	MetodoLogin         string
	Motor               string
	ClaveAdaptador      string
	RequiereLoginGlobal bool
	SoportaLectura      bool
	SoportaAutoregistro bool
	Estado              string
	CreadoEn            time.Time
	CreadoPor           *uuid.UUID
	ActualizadoEn       *time.Time
	ActualizadoPor      *uuid.UUID
}

type ConexionLectura struct {
	Id                uuid.UUID
	SistemaDestinoId  uuid.UUID
	Host              string
	Puerto            int64
	BaseDatos         string
	UsuarioDbCifrado  []byte
	PasswordDbCifrada []byte
	SslModo           string
	ParametrosExtra   []byte
	Estado            string
	CreadoEn          time.Time
	CreadoPor         *uuid.UUID
}

type ParametroHashDestino struct {
	Id               uuid.UUID
	SistemaDestinoId uuid.UUID
	Algoritmo        string
	Costo            *int64
	SalEstrategia    string
	EsDefault        bool
	Notas            string
	CreadoEn         time.Time
	CreadoPor        *uuid.UUID
}
