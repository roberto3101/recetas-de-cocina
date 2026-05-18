package catalogo_sistemas

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosActualizarSistema struct {
	SistemaId           uuid.UUID
	Nombre              string
	UrlAcceso           string
	UrlLogin            string
	NombreCampoUsuario  string
	NombreCampoPassword string
	MetodoLogin         string
	Estado              string
	ActualizadoPor      uuid.UUID
	SesionId            *uuid.UUID
	IpOrigen            string
	AgenteUsuario       string
}

func ActualizarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosActualizarSistema) error {
	existente, err := ConsultarSistemaPorIdDb(contexto, conexion.Pool(), datos.SistemaId)
	if err != nil {
		return err
	}

	if strings.TrimSpace(datos.Nombre) == "" {
		return ErrNombreRequerido
	}
	if strings.TrimSpace(datos.UrlAcceso) == "" {
		return ErrUrlAccesoRequerida
	}

	existente.Nombre = strings.TrimSpace(datos.Nombre)
	existente.UrlAcceso = strings.TrimSpace(datos.UrlAcceso)
	if strings.TrimSpace(datos.UrlLogin) != "" {
		existente.UrlLogin = strings.TrimSpace(datos.UrlLogin)
	}
	if strings.TrimSpace(datos.NombreCampoUsuario) != "" {
		existente.NombreCampoUsuario = strings.TrimSpace(datos.NombreCampoUsuario)
	}
	if strings.TrimSpace(datos.NombreCampoPassword) != "" {
		existente.NombreCampoPassword = strings.TrimSpace(datos.NombreCampoPassword)
	}
	if datos.MetodoLogin != "" {
		existente.MetodoLogin = datos.MetodoLogin
	}
	if datos.Estado != "" {
		existente.Estado = datos.Estado
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := ActualizarSistemaDb(contexto, trans, existente, datos.ActualizadoPor); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &datos.ActualizadoPor,
		SesionId:      datos.SesionId,
		Modulo:        "CATALOGO_SISTEMAS",
		Accion:        "SISTEMA_ACTUALIZADO",
		Entidad:       "sistema_destino",
		EntidadId:     &existente.Id,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
