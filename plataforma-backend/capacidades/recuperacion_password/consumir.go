package recuperacion_password

import (
	"context"
	"time"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/compartido/validaciones"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosConsumo struct {
	TokenPlano      string
	NuevaPassword   string
	IpOrigen        string
	AgenteUsuario   string
}

// Consumir intercambia un token válido por un cambio de password.
//
// Pasos en orden estricto:
//   1. Validar fortaleza de la nueva password (espejo del backend normal)
//   2. Buscar token por hash; rechazar si no existe / expiró / consumido
//   3. Marcar el token como consumido (operación atómica vía SQL WHERE)
//   4. Actualizar el password_hash del usuario
//   5. Invalidar TODAS las sesiones activas del usuario (lo importante)
//   6. Auditar la operación
//
// El paso 3 antes que 4 es deliberado: si dos peticiones simultáneas con el
// mismo token llegan, solo una pasa el UPDATE de consumido_en (RowsAffected=1)
// y la otra obtiene ErrTokenInvalido. Cierra la race condition.
func Consumir(
	contexto context.Context,
	conexion *cockroach.ConexionBaseDatos,
	datos DatosConsumo,
) error {
	if err := validaciones.ValidarFortalezaPassword(datos.NuevaPassword); err != nil {
		return err
	}

	hash := cripto.HashearTokenConSha256(datos.TokenPlano)
	tok, err := ConsultarPorHash(contexto, conexion.Pool(), hash)
	if err != nil {
		return err
	}
	ahora := time.Now()
	if !tok.EstaVigente(ahora) {
		return ErrTokenInvalido
	}

	nuevoHash, err := cripto.HashearPasswordConArgon2id(datos.NuevaPassword)
	if err != nil {
		return err
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	// Atomic compare-and-set: solo una transacción gana esta carrera.
	if err := MarcarComoConsumido(contexto, trans, tok.Id, ahora); err != nil {
		return err
	}

	if err := identidad.ActualizarPasswordUsuario(contexto, trans, tok.UsuarioId, nuevoHash, tok.UsuarioId); err != nil {
		return err
	}

	// Si el usuario estaba bloqueado por brute force, además limpiamos eso
	// — la recuperación es prueba de identidad suficiente.
	if err := identidad.DesbloquearUsuario(contexto, trans, tok.UsuarioId, tok.UsuarioId); err != nil {
		// No-op si no estaba BLOQUEADO; un error real igual se propaga.
		return err
	}

	if err := identidad.InvalidarTodasLasSesionesDeUsuario(contexto, trans, tok.UsuarioId, "recuperacion de password"); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &tok.UsuarioId,
		Modulo:        "RECUPERACION",
		Accion:        "PASSWORD_RESTABLECIDO",
		Entidad:       "usuario",
		EntidadId:     &tok.UsuarioId,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	return trans.Commit(contexto)
}
