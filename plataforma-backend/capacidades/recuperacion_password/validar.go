package recuperacion_password

import (
	"context"
	"time"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

// ResultadoValidacion: lo mínimo que el frontend necesita para mostrar la
// pantalla de "ingresa tu nueva contraseña" sin revelar de qué usuario es.
// Devolvemos el correo enmascarado para que el usuario confirme visualmente
// que está reseteando la cuenta correcta sin que el atacante con el link
// vea el correo completo de la víctima.
type ResultadoValidacion struct {
	CorreoEnmascarado string
}

// Validar comprueba que el token existe, no expiró y no fue consumido.
// NO marca el token como consumido — solo se consume al cambiar la password.
// Esto permite que la página de "ingresar nueva clave" pueda re-cargarse
// sin invalidar el token.
func Validar(
	contexto context.Context,
	conexion *cockroach.ConexionBaseDatos,
	tokenPlano string,
) (*ResultadoValidacion, error) {
	hash := cripto.HashearTokenConSha256(tokenPlano)
	tok, err := ConsultarPorHash(contexto, conexion.Pool(), hash)
	if err != nil {
		return nil, err
	}
	if !tok.EstaVigente(time.Now()) {
		return nil, ErrTokenInvalido
	}

	usuario, err := identidad.ConsultarUsuarioPorId(contexto, conexion.Pool(), tok.UsuarioId)
	if err != nil {
		return nil, ErrTokenInvalido
	}

	return &ResultadoValidacion{
		CorreoEnmascarado: enmascararCorreo(usuario.CorreoElectronico),
	}, nil
}

// enmascararCorreo convierte "ejemplo@dominio.com" en "ej****@dominio.com".
// Suficiente para que el dueño reconozca la cuenta sin filtrar al atacante.
func enmascararCorreo(correo string) string {
	arroba := indexarArroba(correo)
	if arroba <= 2 {
		return correo // demasiado corto para enmascarar útilmente
	}
	return correo[:2] + "****" + correo[arroba:]
}

func indexarArroba(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return i
		}
	}
	return -1
}
