package auditoria

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EntradaAuditoria struct {
	UsuarioId       *uuid.UUID
	SesionId        *uuid.UUID
	Modulo          string
	Accion          string
	Entidad         string
	EntidadId       *uuid.UUID
	DatosAnteriores any
	DatosNuevos     any
	Detalle         string
	IpOrigen        string
	AgenteUsuario   string
}

func RegistrarAccion(contexto context.Context, transaccion pgx.Tx, entrada EntradaAuditoria) error {
	var anterioresJson, nuevosJson []byte
	var err error

	if entrada.DatosAnteriores != nil {
		anterioresJson, err = json.Marshal(entrada.DatosAnteriores)
		if err != nil {
			return err
		}
	}
	if entrada.DatosNuevos != nil {
		nuevosJson, err = json.Marshal(entrada.DatosNuevos)
		if err != nil {
			return err
		}
	}

	_, err = transaccion.Exec(contexto, `
		INSERT INTO auditoria_accion (
			usuario_id, sesion_id, modulo, accion,
			entidad, entidad_id,
			datos_anteriores, datos_nuevos,
			detalle, ip_origen, agente_usuario
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		entrada.UsuarioId,
		entrada.SesionId,
		entrada.Modulo,
		entrada.Accion,
		nullableTexto(entrada.Entidad),
		entrada.EntidadId,
		nullableJson(anterioresJson),
		nullableJson(nuevosJson),
		nullableTexto(entrada.Detalle),
		nullableTexto(entrada.IpOrigen),
		nullableTexto(entrada.AgenteUsuario),
	)

	if err != nil {
		slog.Error("no se pudo registrar acción en auditoría",
			"modulo", entrada.Modulo,
			"accion", entrada.Accion,
			"detalle", err.Error())
	}
	return err
}

func nullableTexto(valor string) any {
	if valor == "" {
		return nil
	}
	return valor
}

func nullableJson(valor []byte) any {
	if len(valor) == 0 {
		return nil
	}
	return valor
}
