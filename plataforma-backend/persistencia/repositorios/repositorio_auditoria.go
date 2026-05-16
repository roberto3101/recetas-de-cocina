package repositorios

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sistemas-unificados/persistencia/cockroach"
)

type RegistroAuditoria struct {
	Id              uuid.UUID
	UsuarioId       *uuid.UUID
	SesionId        *uuid.UUID
	Modulo          string
	Accion          string
	Entidad         string
	EntidadId       *uuid.UUID
	Detalle         string
	IpOrigen        string
	AgenteUsuario   string
	CreadoEn        time.Time
}

type FiltroAuditoria struct {
	Modulo       string
	Accion       string
	UsuarioId    *uuid.UUID
	DesdeFecha   *time.Time
	HastaFecha   *time.Time
	Pagina       int
	TamanoPagina int
}

type ResultadoAuditoria struct {
	Registros      []RegistroAuditoria
	TotalRegistros int64
	Pagina         int
	TamanoPagina   int
}

func ConsultarAuditoria(contexto context.Context, ejecutor cockroach.EjecutorSql, filtro FiltroAuditoria) (*ResultadoAuditoria, error) {
	if filtro.TamanoPagina <= 0 {
		filtro.TamanoPagina = 100
	}
	if filtro.TamanoPagina > 200 {
		filtro.TamanoPagina = 200
	}
	if filtro.Pagina < 1 {
		filtro.Pagina = 1
	}
	desplazamiento := (filtro.Pagina - 1) * filtro.TamanoPagina

	var condiciones []string
	var argumentos []any
	indice := 1

	if filtro.Modulo != "" {
		condiciones = append(condiciones, fmt.Sprintf("modulo = $%d", indice))
		argumentos = append(argumentos, filtro.Modulo)
		indice++
	}
	if filtro.Accion != "" {
		condiciones = append(condiciones, fmt.Sprintf("accion = $%d", indice))
		argumentos = append(argumentos, filtro.Accion)
		indice++
	}
	if filtro.UsuarioId != nil {
		condiciones = append(condiciones, fmt.Sprintf("usuario_id = $%d", indice))
		argumentos = append(argumentos, *filtro.UsuarioId)
		indice++
	}
	if filtro.DesdeFecha != nil {
		condiciones = append(condiciones, fmt.Sprintf("creado_en >= $%d", indice))
		argumentos = append(argumentos, *filtro.DesdeFecha)
		indice++
	}
	if filtro.HastaFecha != nil {
		condiciones = append(condiciones, fmt.Sprintf("creado_en <= $%d", indice))
		argumentos = append(argumentos, *filtro.HastaFecha)
		indice++
	}

	donde := ""
	if len(condiciones) > 0 {
		donde = "WHERE " + strings.Join(condiciones, " AND ")
	}

	var total int64
	if err := ejecutor.QueryRow(contexto,
		"SELECT count(*) FROM auditoria_accion "+donde, argumentos...,
	).Scan(&total); err != nil {
		return nil, err
	}

	consulta := `
		SELECT id, usuario_id, sesion_id, modulo, accion,
		       coalesce(entidad,''), entidad_id,
		       coalesce(detalle,''), coalesce(ip_origen,''), coalesce(agente_usuario,''),
		       creado_en
		FROM auditoria_accion ` + donde + fmt.Sprintf(`
		ORDER BY creado_en DESC
		LIMIT $%d OFFSET $%d
	`, indice, indice+1)
	argumentos = append(argumentos, filtro.TamanoPagina, desplazamiento)

	filas, err := ejecutor.Query(contexto, consulta, argumentos...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	resultados := make([]RegistroAuditoria, 0)
	for filas.Next() {
		r := RegistroAuditoria{}
		if err := filas.Scan(
			&r.Id, &r.UsuarioId, &r.SesionId, &r.Modulo, &r.Accion,
			&r.Entidad, &r.EntidadId, &r.Detalle, &r.IpOrigen, &r.AgenteUsuario,
			&r.CreadoEn,
		); err != nil {
			return nil, err
		}
		resultados = append(resultados, r)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	return &ResultadoAuditoria{
		Registros:      resultados,
		TotalRegistros: total,
		Pagina:         filtro.Pagina,
		TamanoPagina:   filtro.TamanoPagina,
	}, nil
}
