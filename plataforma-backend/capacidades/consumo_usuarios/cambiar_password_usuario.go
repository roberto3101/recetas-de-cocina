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

type DatosCambiarPasswordUsuarioExterno struct {
	SistemaId     uuid.UUID
	IdExterno     string
	PasswordNueva string
	OperadorId    uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

func CambiarPasswordUsuarioEnSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, clavesCifrado *cripto.ClavesCifrado, datos DatosCambiarPasswordUsuarioExterno) error {
	sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(contexto, conexion.Pool(), datos.SistemaId)
	if err != nil {
		return err
	}
	if sistema.Estado != catalogo_sistemas.EstadoSistemaActivo {
		return errors.New("el sistema no está activo")
	}
	adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
	if err != nil {
		return err
	}

	if err := adaptador.CambiarPasswordUsuario(contexto, datos.IdExterno, adaptadores.DatosCambiarPasswordExterno{
		PasswordNueva: datos.PasswordNueva,
	}); err != nil {
		return err
	}

	passwordCifrada, err := cripto.CifrarConAesGcm(clavesCifrado.ClaveBoveda(), []byte(datos.PasswordNueva))
	if err != nil {
		return err
	}
	credencial := &catalogo_sistemas.CredencialUsuarioExterno{
		SistemaDestinoId: sistema.Id,
		IdExternoUsuario: datos.IdExterno,
		PasswordCifrada:  passwordCifrada,
		Estado:           "ACTIVO",
		CreadoPor:        &datos.OperadorId,
	}
	if err := catalogo_sistemas.GuardarCredencialUsuarioExterno(contexto, conexion.Pool(), credencial); err != nil {
		return err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.OperadorId,
		SesionId:  datos.SesionId,
		Modulo:    "CONSUMO_USUARIOS",
		Accion:    "USUARIO_EXTERNO_PASSWORD_CAMBIADA",
		Entidad:   "sistema_destino",
		EntidadId: &sistema.Id,
		DatosNuevos: map[string]any{
			"sistema_codigo": sistema.Codigo,
			"id_externo":     datos.IdExterno,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}
	return trans.Commit(contexto)
}
