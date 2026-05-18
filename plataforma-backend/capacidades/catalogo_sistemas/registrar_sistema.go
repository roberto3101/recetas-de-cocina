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
	UrlAcceso           string
	UrlLogin            string
	NombreCampoUsuario  string
	NombreCampoPassword string
	MetodoLogin         string

	CreadoPor     uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

type ResultadoRegistrarSistema struct {
	SistemaId uuid.UUID
}

func RegistrarSistema(contexto context.Context, conexion *cockroach.ConexionBaseDatos, _ *cripto.ClavesCifrado, datos DatosRegistrarSistema) (*ResultadoRegistrarSistema, error) {
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

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	urlLogin := strings.TrimSpace(datos.UrlLogin)
	if urlLogin == "" {
		urlLogin = strings.TrimSpace(datos.UrlAcceso)
	}
	metodoLogin := datos.MetodoLogin
	if metodoLogin == "" {
		metodoLogin = MetodoLoginPost
	}

	sistema := &SistemaDestino{
		Codigo:              codigo,
		Nombre:              strings.TrimSpace(datos.Nombre),
		UrlAcceso:           strings.TrimSpace(datos.UrlAcceso),
		UrlLogin:            urlLogin,
		NombreCampoUsuario:  strings.TrimSpace(datos.NombreCampoUsuario),
		NombreCampoPassword: strings.TrimSpace(datos.NombreCampoPassword),
		MetodoLogin:         metodoLogin,
		Estado:              EstadoSistemaActivo,
		CreadoPor:           &datos.CreadoPor,
	}

	if err := InsertarSistemaDestino(contexto, trans, sistema); err != nil {
		if ErrorEsDuplicadoCodigoSistema(err) {
			return nil, ErrCodigoSistemaDuplicado
		}
		return nil, err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId: &datos.CreadoPor,
		SesionId:  datos.SesionId,
		Modulo:    "CATALOGO_SISTEMAS",
		Accion:    "SISTEMA_REGISTRADO",
		Entidad:   "sistema_destino",
		EntidadId: &sistema.Id,
		DatosNuevos: map[string]any{
			"codigo":     codigo,
			"nombre":     sistema.Nombre,
			"url_acceso": sistema.UrlAcceso,
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
