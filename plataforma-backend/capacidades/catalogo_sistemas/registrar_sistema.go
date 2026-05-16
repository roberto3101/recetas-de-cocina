package catalogo_sistemas

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosRegistrarSistema struct {
	Codigo              string
	Nombre              string
	Descripcion         string
	UrlAcceso           string
	Motor               string
	ClaveAdaptador      string
	RequiereLoginGlobal bool
	SoportaLectura      bool
	SoportaAutoregistro bool

	HostLectura          string
	PuertoLectura        int64
	BaseDatosLectura     string
	UsuarioLecturaPlano  string
	PasswordLecturaPlana string
	SslModoLectura       string
	ParametrosExtraJson  []byte

	AlgoritmoHashDestino string
	CostoHashDestino     *int64
	SalEstrategia        string

	CreadoPor     uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

type ResultadoRegistrarSistema struct {
	SistemaId uuid.UUID
}

func RegistrarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado, datos DatosRegistrarSistema) (*ResultadoRegistrarSistema, error) {
	codigo := strings.ToLower(strings.TrimSpace(datos.Codigo))
	if codigo == "" {
		return nil, ErrCodigoRequerido
	}
	if strings.TrimSpace(datos.Nombre) == "" {
		return nil, ErrNombreRequerido
	}
	if strings.TrimSpace(datos.UrlAcceso) == "" {
		return nil, ErrUrlAccesoRequerida
	}
	if !esMotorValido(datos.Motor) {
		return nil, ErrMotorInvalido
	}
	if strings.TrimSpace(datos.ClaveAdaptador) == "" {
		return nil, ErrAdaptadorRequerido
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	sistema := &SistemaDestino{
		Codigo:              codigo,
		Nombre:              strings.TrimSpace(datos.Nombre),
		Descripcion:         strings.TrimSpace(datos.Descripcion),
		UrlAcceso:           strings.TrimSpace(datos.UrlAcceso),
		Motor:               datos.Motor,
		ClaveAdaptador:      strings.TrimSpace(datos.ClaveAdaptador),
		RequiereLoginGlobal: datos.RequiereLoginGlobal,
		SoportaLectura:      datos.SoportaLectura,
		SoportaAutoregistro: datos.SoportaAutoregistro,
		Estado:              EstadoSistemaActivo,
		CreadoPor:           &datos.CreadoPor,
	}

	if err := InsertarSistemaDestino(contexto, trans, sistema); err != nil {
		if ErrorEsDuplicadoCodigoSistema(err) {
			return nil, ErrCodigoSistemaDuplicado
		}
		return nil, err
	}

	if datos.HostLectura != "" {
		clave := clavesCifrado.ClaveBoveda()
		usuarioDbCifrado, err := cripto.CifrarConAesGcm(clave, []byte(datos.UsuarioLecturaPlano))
		if err != nil {
			return nil, err
		}
		passwordDbCifrada, err := cripto.CifrarConAesGcm(clave, []byte(datos.PasswordLecturaPlana))
		if err != nil {
			return nil, err
		}

		sslModo := datos.SslModoLectura
		if sslModo == "" {
			sslModo = "require"
		}

		conexionLectura := &ConexionLectura{
			SistemaDestinoId:  sistema.Id,
			Host:              datos.HostLectura,
			Puerto:            datos.PuertoLectura,
			BaseDatos:         datos.BaseDatosLectura,
			UsuarioDbCifrado:  usuarioDbCifrado,
			PasswordDbCifrada: passwordDbCifrada,
			SslModo:           sslModo,
			ParametrosExtra:   datos.ParametrosExtraJson,
			Estado:            EstadoSistemaActivo,
			CreadoPor:         &datos.CreadoPor,
		}

		if err := InsertarConexionLectura(contexto, trans, conexionLectura); err != nil {
			return nil, err
		}
	}

	if datos.AlgoritmoHashDestino != "" {
		salEstrategia := datos.SalEstrategia
		if salEstrategia == "" {
			salEstrategia = "POR_FILA"
		}
		parametro := &ParametroHashDestino{
			SistemaDestinoId: sistema.Id,
			Algoritmo:        datos.AlgoritmoHashDestino,
			Costo:            datos.CostoHashDestino,
			SalEstrategia:    salEstrategia,
			EsDefault:        true,
			CreadoPor:        &datos.CreadoPor,
		}
		if err := InsertarParametroHash(contexto, trans, parametro); err != nil {
			return nil, err
		}
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.CreadoPor,
		SesionId:  datos.SesionId,
		Modulo:    "CATALOGO_SISTEMAS",
		Accion:    "SISTEMA_REGISTRADO",
		Entidad:   "sistema_destino",
		EntidadId: &sistema.Id,
		DatosNuevos: map[string]any{
			"codigo":    codigo,
			"motor":     datos.Motor,
			"adaptador": datos.ClaveAdaptador,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return nil, err
	}

	if err := trans.Commit(contexto); err != nil {
		return nil, err
	}

	return &ResultadoRegistrarSistema{SistemaId: sistema.Id}, nil
}

func esMotorValido(motor string) bool {
	switch motor {
	case MotorCockroachdb, MotorPostgresql, MotorMysql, MotorMariadb, MotorApiRest, MotorOtro:
		return true
	}
	return false
}
