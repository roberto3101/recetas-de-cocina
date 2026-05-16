package consumo_usuarios

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

var ErrUsuarioExternoNoEncontradoPorCorreo = errors.New("no se encontró un usuario en el sistema externo con ese correo")

type DatosAsociarCredencialExterna struct {
	SistemaId     uuid.UUID
	Correo        string
	PasswordPlana string
	OperadorId    uuid.UUID
	SesionId      *uuid.UUID
	IpOrigen      string
	AgenteUsuario string
}

type ResultadoAsociacion struct {
	IdExterno         string `json:"id_externo"`
	CorreoElectronico string `json:"correo_electronico"`
	SistemaCodigo     string `json:"sistema_codigo"`
	SistemaNombre     string `json:"sistema_nombre"`
}

func AsociarCredencialExterna(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, clavesCifrado *cripto.ClavesCifrado, datos DatosAsociarCredencialExterna) (*ResultadoAsociacion, error) {
	correo := strings.TrimSpace(strings.ToLower(datos.Correo))
	if correo == "" {
		return nil, errors.New("correo requerido")
	}
	if datos.PasswordPlana == "" {
		return nil, errors.New("password requerida")
	}

	sistema, err := catalogo_sistemas.ConsultarSistemaPorIdDb(contexto, conexion.Pool(), datos.SistemaId)
	if err != nil {
		return nil, err
	}
	if sistema.Estado != catalogo_sistemas.EstadoSistemaActivo {
		return nil, errors.New("el sistema no está activo")
	}
	if !sistema.SoportaLectura {
		return nil, errors.New("el sistema no soporta lectura: no se puede verificar el usuario")
	}

	adaptador, err := resolver.Resolver(sistema.ClaveAdaptador, sistema.Id)
	if err != nil {
		return nil, err
	}

	listado, err := adaptador.ListarUsuarios(contexto, adaptadores.FiltroBusquedaUsuarios{
		Texto:        correo,
		Pagina:       1,
		TamanoPagina: 50,
	})
	if err != nil {
		return nil, err
	}

	var encontrado *adaptadores.UsuarioExterno
	for i := range listado.Usuarios {
		if strings.EqualFold(strings.TrimSpace(listado.Usuarios[i].CorreoElectronico), correo) {
			encontrado = &listado.Usuarios[i]
			break
		}
	}
	if encontrado == nil {
		return nil, ErrUsuarioExternoNoEncontradoPorCorreo
	}

	passwordCifrada, err := cripto.CifrarConAesGcm(clavesCifrado.ClaveBoveda(), []byte(datos.PasswordPlana))
	if err != nil {
		return nil, err
	}

	credencial := &catalogo_sistemas.CredencialUsuarioExterno{
		SistemaDestinoId: sistema.Id,
		IdExternoUsuario: encontrado.IdExterno,
		CorreoUsuario:    encontrado.CorreoElectronico,
		PasswordCifrada:  passwordCifrada,
		Estado:           "ACTIVO",
		CreadoPor:        &datos.OperadorId,
	}
	if err := catalogo_sistemas.GuardarCredencialUsuarioExterno(contexto, conexion.Pool(), credencial); err != nil {
		return nil, err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err == nil {
		_ = auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
			UsuarioId: &datos.OperadorId,
			SesionId:  datos.SesionId,
			Modulo:    "CONSUMO_USUARIOS",
			Accion:    "CREDENCIAL_EXTERNA_ASOCIADA",
			Entidad:   "sistema_destino",
			EntidadId: &sistema.Id,
			DatosNuevos: map[string]any{
				"sistema_codigo": sistema.Codigo,
				"correo":         encontrado.CorreoElectronico,
				"id_externo":     encontrado.IdExterno,
			},
			IpOrigen:      datos.IpOrigen,
			AgenteUsuario: datos.AgenteUsuario,
		})
		_ = trans.Commit(contexto)
	}

	return &ResultadoAsociacion{
		IdExterno:         encontrado.IdExterno,
		CorreoElectronico: encontrado.CorreoElectronico,
		SistemaCodigo:     sistema.Codigo,
		SistemaNombre:     sistema.Nombre,
	}, nil
}
