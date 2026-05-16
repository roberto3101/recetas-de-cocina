package identidad

import (
	"context"

	"github.com/google/uuid"

	"sistemas-unificados/compartido/validaciones"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosCambioPassword struct {
	UsuarioId      uuid.UUID
	PasswordActual string
	PasswordNueva  string
	SesionId       *uuid.UUID
	IpOrigen       string
	AgenteUsuario  string
}

func CambiarPassword(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosCambioPassword) error {
	if err := validaciones.ValidarFortalezaPassword(datos.PasswordNueva); err != nil {
		return err
	}

	usuarioActual, err := ConsultarUsuarioPorId(contexto, conexion.Pool(), datos.UsuarioId)
	if err != nil {
		return err
	}
	if usuarioActual.Estado == EstadoUsuarioEliminado {
		return ErrUsuarioEliminado
	}

	coincide, err := cripto.VerificarPasswordContraHashArgon2id(datos.PasswordActual, usuarioActual.PasswordHash)
	if err != nil {
		return err
	}
	if !coincide {
		return ErrPasswordActualInvalido
	}

	nuevoHash, err := cripto.HashearPasswordConArgon2id(datos.PasswordNueva)
	if err != nil {
		return err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := ActualizarPasswordUsuario(contexto, trans, datos.UsuarioId, nuevoHash, datos.UsuarioId); err != nil {
		return err
	}

	if err := InvalidarTodasLasSesionesDeUsuario(contexto, trans, datos.UsuarioId, "cambio de password"); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.UsuarioId,
		SesionId:      datos.SesionId,
		Modulo:        "IDENTIDAD",
		Accion:        "PASSWORD_ACTUALIZADO",
		Entidad:       "usuario",
		EntidadId:     &datos.UsuarioId,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
