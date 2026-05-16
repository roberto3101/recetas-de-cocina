package adaptadores

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrOperacionNoSoportada = errors.New("operación no soportada por este adaptador")
var ErrUsuarioExternoNoEncontrado = errors.New("usuario no encontrado en el sistema externo")

type UsuarioExterno struct {
	IdExterno         string
	CorreoElectronico string
	NombreCompleto    string
	Estado            string
	PasswordHash      string
	DatosAdicionales  map[string]string
}

type FiltroBusquedaUsuarios struct {
	Texto        string
	Estado       string
	Pagina       int
	TamanoPagina int
}

type ResultadoBusquedaUsuarios struct {
	Usuarios       []UsuarioExterno
	TotalRegistros int64
	Pagina         int
	TamanoPagina   int
}

type DatosProvisionar struct {
	CorreoElectronico string
	PasswordPlana     string
	NombreCompleto    string
}

type DatosEditarUsuarioExterno struct {
	CorreoElectronicoNuevo string
	NombreCompletoNuevo    string
}

type DatosCambiarPasswordExterno struct {
	PasswordNueva string
}

type TablaConCount struct {
	Schema string `json:"schema"`
	Nombre string `json:"nombre"`
	Filas  int64  `json:"filas"`
}

type Diagnostico struct {
	BaseDatos string          `json:"base_datos"`
	Tablas    []TablaConCount `json:"tablas"`
}

type Adaptador interface {
	Clave() string
	ListarUsuarios(contexto context.Context, filtro FiltroBusquedaUsuarios) (*ResultadoBusquedaUsuarios, error)
	ConsultarUsuario(contexto context.Context, idExterno string) (*UsuarioExterno, error)
	ProvisionarUsuario(contexto context.Context, datos DatosProvisionar) (*UsuarioExterno, error)
	EditarUsuario(contexto context.Context, idExterno string, datos DatosEditarUsuarioExterno) (*UsuarioExterno, error)
	CambiarPasswordUsuario(contexto context.Context, idExterno string, datos DatosCambiarPasswordExterno) error
	ProbarConexion(contexto context.Context) error
	Diagnosticar(contexto context.Context) (*Diagnostico, error)
}

type Resolver interface {
	Resolver(claveAdaptador string, idSistemaDestino uuid.UUID) (Adaptador, error)
}
