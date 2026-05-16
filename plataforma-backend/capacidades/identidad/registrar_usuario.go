package identidad

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/compartido/validaciones"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosRegistroUsuario struct {
	CorreoElectronico string
	PasswordPlana     string
	RegistradoPor     *uuid.UUID
	IpOrigen          string
	AgenteUsuario     string
}

type ResultadoRegistroUsuario struct {
	UsuarioId uuid.UUID
	Estado    string
}

func RegistrarUsuario(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosRegistroUsuario) (*ResultadoRegistroUsuario, error) {
	correoLimpio := validaciones.NormalizarCorreo(datos.CorreoElectronico)
	if !validaciones.EsCorreoElectronicoValido(correoLimpio) {
		return nil, ErrCorreoInvalido
	}
	if err := validaciones.ValidarFortalezaPassword(datos.PasswordPlana); err != nil {
		return nil, err
	}

	existente, err := ConsultarUsuarioPorCorreo(contexto, conexion.Pool(), correoLimpio)
	if err != nil && !errors.Is(err, ErrUsuarioNoEncontrado) {
		return nil, err
	}
	if existente != nil {
		return nil, ErrCorreoYaRegistrado
	}

	hash, err := cripto.HashearPasswordConArgon2id(datos.PasswordPlana)
	if err != nil {
		return nil, err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	nuevo := &Usuario{
		CorreoElectronico: correoLimpio,
		PasswordHash:      hash,
		Estado:            EstadoUsuarioPendiente,
		CreadoPor:         datos.RegistradoPor,
	}

	if err := InsertarUsuario(contexto, trans, nuevo); err != nil {
		return nil, err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: datos.RegistradoPor,
		Modulo:    "IDENTIDAD",
		Accion:    "USUARIO_REGISTRADO",
		Entidad:   "usuario",
		EntidadId: &nuevo.Id,
		DatosNuevos: map[string]any{
			"correo_electronico": correoLimpio,
			"estado":             nuevo.Estado,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return nil, err
	}

	if err := trans.Commit(contexto); err != nil {
		return nil, err
	}

	return &ResultadoRegistroUsuario{
		UsuarioId: nuevo.Id,
		Estado:    nuevo.Estado,
	}, nil
}

func RegistrarUsuarioYActivar(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosRegistroUsuario) (*ResultadoRegistroUsuario, error) {
	resultado, err := RegistrarUsuario(contexto, conexion, datos)
	if err != nil {
		return nil, err
	}
	if err := MarcarCorreoVerificadoYActivar(contexto, conexion.Pool(), resultado.UsuarioId); err != nil {
		return nil, err
	}
	resultado.Estado = EstadoUsuarioActivo
	return resultado, nil
}
