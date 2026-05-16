package consumo_usuarios

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosEditarUsuarioExterno struct {
	SistemaId              uuid.UUID
	IdExterno              string
	CorreoElectronicoNuevo string
	NombreCompletoNuevo    string
	OperadorId             uuid.UUID
	SesionId               *uuid.UUID
	IpOrigen               string
	AgenteUsuario          string
}

func EditarUsuarioEnSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, datos DatosEditarUsuarioExterno) (*adaptadores.UsuarioExterno, error) {
	sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(contexto, conexion.Pool(), datos.SistemaId)
	if err != nil {
		return nil, err
	}
	if sistema.Estado != catalogo_sistemas.EstadoSistemaActivo {
		return nil, errors.New("el sistema no está activo")
	}
	adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
	if err != nil {
		return nil, err
	}

	usuario, err := adaptador.EditarUsuario(contexto, datos.IdExterno, adaptadores.DatosEditarUsuarioExterno{
		CorreoElectronicoNuevo: datos.CorreoElectronicoNuevo,
		NombreCompletoNuevo:    datos.NombreCompletoNuevo,
	})
	if err != nil {
		return nil, err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return usuario, err
	}
	defer trans.Rollback(contexto)

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.OperadorId,
		SesionId:  datos.SesionId,
		Modulo:    "CONSUMO_USUARIOS",
		Accion:    "USUARIO_EXTERNO_EDITADO",
		Entidad:   "sistema_destino",
		EntidadId: &sistema.Id,
		DatosNuevos: map[string]any{
			"sistema_codigo":          sistema.Codigo,
			"id_externo":              datos.IdExterno,
			"correo_electronico_nuevo": datos.CorreoElectronicoNuevo,
			"nombre_completo_nuevo":   datos.NombreCompletoNuevo,
		},
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return usuario, err
	}
	if err := trans.Commit(contexto); err != nil {
		return usuario, err
	}

	return usuario, nil
}
