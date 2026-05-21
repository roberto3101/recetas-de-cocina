package recuperacion_password

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"os"
	"strings"
	"time"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/correo"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

type DatosSolicitud struct {
	CorreoElectronico string
	IpOrigen          string
	AgenteUsuario     string
}

// Solicitar emite un token único para que el operador cambie su contraseña.
//
// Comportamiento clave anti-enumeración:
//   - Si el correo existe → genera token, invalida los previos, loguea link.
//   - Si el correo NO existe → no hace nada visible para el cliente, pero
//     tampoco retorna error. El frontend muestra el mismo mensaje en ambos
//     casos ("Si el correo existe, te enviamos un enlace").
//
// El token plano se devuelve para que el caller pueda loguearlo / mandarlo
// por correo. NUNCA se persiste en claro: solo se guarda sha256(token).
//
// Si la BD no tiene SMTP configurado (caso actual del proyecto), el token
// se loguea con nivel WARN para que sea visible en `pm2 logs` y el operador
// pueda copiar el link al destinatario por canal lateral (WhatsApp/Signal).
func Solicitar(
	contexto context.Context,
	conexion *cockroach.ConexionBaseDatos,
	datos DatosSolicitud,
) error {
	correoLimpio := strings.ToLower(strings.TrimSpace(datos.CorreoElectronico))
	if correoLimpio == "" {
		// No revelamos al cliente que el correo viene vacío — simplemente
		// no hacemos nada (cero side effects).
		return nil
	}

	usuario, err := identidad.ConsultarUsuarioPorCorreo(contexto, conexion.Pool(), correoLimpio)
	if err != nil {
		// ¿Usuario no encontrado? Silencio absoluto. No es error, no es log.
		// Cualquier otra falla técnica sí la propagamos.
		if err.Error() == identidad.ErrUsuarioNoEncontrado.Error() {
			return nil
		}
		return err
	}

	// Solo emitimos tokens para usuarios activos. Pendientes, bloqueados,
	// inactivos o eliminados: no pasa nada (silencio).
	if usuario.Estado != identidad.EstadoUsuarioActivo {
		return nil
	}

	tokenPlano, tokenHash, err := generarTokenAleatorio()
	if err != nil {
		return err
	}

	ahora := time.Now()
	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return err
	}
	defer trans.Rollback(contexto)

	if err := InvalidarTokensVigentesDeUsuario(contexto, trans, usuario.Id, ahora); err != nil {
		return err
	}

	tok := &Token{
		UsuarioId:     usuario.Id,
		TokenHash:     tokenHash,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
		ExpiraEn:      ahora.Add(VigenciaDelToken),
	}
	if err := InsertarToken(contexto, trans, tok); err != nil {
		return err
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &usuario.Id,
		Modulo:        "RECUPERACION",
		Accion:        "TOKEN_EMITIDO",
		Entidad:       "token_recuperacion",
		EntidadId:     &tok.Id,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return err
	}

	if err := trans.Commit(contexto); err != nil {
		return err
	}

	urlBase := strings.TrimSuffix(os.Getenv("URL_PUBLICA_FRONTEND"), "/")
	if urlBase == "" {
		urlBase = "https://recetas-cocina.duckdns.org"
	}
	enlace := urlBase + "/receta/" + tokenPlano

	// Log SIEMPRE — el link en pm2 logs es el fallback si SMTP falla, está
	// caído o el cortocircuito está abierto. El operador siempre puede
	// recuperarlo desde el servidor.
	slog.Warn("recuperacion_password.link_emitido",
		"correo", correoLimpio,
		"enlace", enlace,
		"expira_en", tok.ExpiraEn.Format(time.RFC3339),
		"ip_origen", datos.IpOrigen,
	)

	// Envío del correo: si falla por cualquier motivo, NO propagamos el
	// error. El usuario ve el mismo mensaje "si existe, te llegará" — y el
	// operador tiene el link en logs igual. Esto cumple dos cosas:
	//   1. Anti-enumeración: respuesta idéntica caiga o no el correo.
	//   2. Anti-bloqueo del proveedor: un solo intento, sin reintentos.
	mensaje := correo.MensajeCorreo{
		Destinatario: correoLimpio,
		Asunto:       "Tu receta perdida",
		Cuerpo: "Hola,\r\n\r\n" +
			"Solicitaste reescribir tu receta secreta.\r\n" +
			"Usa este enlace dentro de los próximos 30 minutos:\r\n\r\n" +
			enlace + "\r\n\r\n" +
			"Si no fuiste tú, ignora este mensaje.\r\n\r\n" +
			"— Recetas del Chef\r\n",
	}
	if err := correo.Enviar(contexto, mensaje); err != nil {
		// Ya está logueado dentro de Enviar(). Aquí solo aclaramos que se
		// está usando el fallback de pm2 logs.
		slog.Info("recuperacion_password.fallback_pm2_logs",
			"correo", correoLimpio,
			"razon", err.Error(),
		)
	}

	return nil
}

func generarTokenAleatorio() (string, string, error) {
	bytesAleatorios := make([]byte, LongitudDelToken)
	if _, err := rand.Read(bytesAleatorios); err != nil {
		return "", "", err
	}
	// URL-safe sin padding: queda limpio en la barra de direcciones.
	plano := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytesAleatorios)
	hash := cripto.HashearTokenConSha256(plano)
	return plano, hash, nil
}
