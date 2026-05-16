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
	Descripcion         string
	UrlAcceso           string
	Motor               string
	ClaveAdaptador      string
	RequiereLoginGlobal bool
	SoportaLectura      bool
	SoportaAutoregistro bool
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
	if !esMotorValido(datos.Motor) {
		return ErrMotorInvalido
	}

	existente.Nombre = strings.TrimSpace(datos.Nombre)
	existente.Descripcion = strings.TrimSpace(datos.Descripcion)
	existente.UrlAcceso = strings.TrimSpace(datos.UrlAcceso)
	existente.Motor = datos.Motor
	existente.ClaveAdaptador = strings.TrimSpace(datos.ClaveAdaptador)
	existente.RequiereLoginGlobal = datos.RequiereLoginGlobal
	existente.SoportaLectura = datos.SoportaLectura
	existente.SoportaAutoregistro = datos.SoportaAutoregistro
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
