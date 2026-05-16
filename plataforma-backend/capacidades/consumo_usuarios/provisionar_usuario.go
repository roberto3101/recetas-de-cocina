package consumo_usuarios

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosProvisionarUsuarioExterno struct {
	SistemaId         uuid.UUID
	CorreoElectronico string
	PasswordPlana     string
	NombreCompleto    string
	OperadorId        uuid.UUID
	SesionId          *uuid.UUID
	IpOrigen          string
	AgenteUsuario     string
}

func ProvisionarUsuarioEnSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, clavesCifrado *cripto.ClavesCifrado, datos DatosProvisionarUsuarioExterno) (*adaptadores.UsuarioExterno, error) {
	sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(contexto, conexion.Pool(), datos.SistemaId)
	if err != nil {
		return nil, err
	}
	if sistema.Estado != catalogo_sistemas.EstadoSistemaActivo {
		return nil, errors.New("el sistema no está activo")
	}
	if !sistema.SoportaAutoregistro {
		return nil, errors.New("el sistema no permite autoregistro de usuarios")
	}

	adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
	if err != nil {
		return nil, err
	}

	usuarioCreado, err := adaptador.ProvisionarUsuario(contexto, adaptadores.DatosProvisionar{
		CorreoElectronico: datos.CorreoElectronico,
		PasswordPlana:     datos.PasswordPlana,
		NombreCompleto:    datos.NombreCompleto,
	})
	if err != nil {
		return nil, err
	}

	passwordCifrada, err := cripto.CifrarConAesGcm(clavesCifrado.ClaveBoveda(), []byte(datos.PasswordPlana))
	if err != nil {
		return usuarioCreado, err
	}

	credencial := &catalogo_sistemas.CredencialUsuarioExterno{
		SistemaDestinoId: sistema.Id,
		IdExternoUsuario: usuarioCreado.IdExterno,
		CorreoUsuario:    usuarioCreado.CorreoElectronico,
		PasswordCifrada:  passwordCifrada,
		Estado:           "ACTIVO",
		CreadoPor:        &datos.OperadorId,
	}
	if err := catalogo_sistemas.GuardarCredencialUsuarioExterno(contexto, conexion.Pool(), credencial); err != nil {
		return usuarioCreado, err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return usuarioCreado, err
	}
	defer trans.Rollback(contexto)

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.OperadorId,
		SesionId:  datos.SesionId,
		Modulo:    "CONSUMO_USUARIOS",
		Accion:    "USUARIO_EXTERNO_PROVISIONADO",
		Entidad:   "sistema_destino",
		EntidadId: &sistema.Id,
		DatosNuevos: map[string]any{
			"sistema_codigo":     sistema.Codigo,
			"correo_electronico": datos.CorreoElectronico,
			"id_externo":         usuarioCreado.IdExterno,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return usuarioCreado, err
	}

	if err := trans.Commit(contexto); err != nil {
		return usuarioCreado, err
	}

	return usuarioCreado, nil
}
