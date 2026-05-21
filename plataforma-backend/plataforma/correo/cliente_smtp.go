package correo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Tiempos cortos: el endpoint que dispara el envío debe responder rápido
// al usuario; si SMTP se cuelga, preferimos abandonar y dejar el link en
// logs como fallback. NUNCA reintentar — el proveedor bloquea a 5 fallos.
const (
	timeoutDial   = 8 * time.Second
	timeoutEnvio  = 12 * time.Second
	limiteErrores = 3                // pausa antes de llegar al límite del proveedor (5)
	duracionPausa = 10 * time.Minute // ventana para que se "enfríe" la cuenta
)

// ConfigSmtp viene de env vars (SMTP_HOST / SMTP_PORT / SMTP_USER / SMTP_PASS).
// Si alguna falta, el cliente queda deshabilitado y todo envío retorna error
// sin tocar la red. Esto permite ejecutar el backend sin SMTP configurado.
type ConfigSmtp struct {
	Host       string
	Puerto     int
	Usuario    string
	Password   string
	Remitente  string // "Nombre <email>" o solo email
	Habilitado bool
}

var (
	configCargada  ConfigSmtp
	configUnaVez   sync.Once
	cortocircuitoGlobal = NuevoCortocircuito(limiteErrores, duracionPausa)
)

// Configuracion carga (una sola vez) y cachea los valores del entorno.
func Configuracion() ConfigSmtp {
	configUnaVez.Do(func() {
		host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
		usuario := strings.TrimSpace(os.Getenv("SMTP_USER"))
		password := os.Getenv("SMTP_PASS")
		puertoStr := strings.TrimSpace(os.Getenv("SMTP_PORT"))

		if host == "" || usuario == "" || password == "" || puertoStr == "" {
			slog.Warn("smtp.deshabilitado", "razon", "faltan variables SMTP_*")
			return
		}
		puerto, err := strconv.Atoi(puertoStr)
		if err != nil || puerto <= 0 {
			slog.Warn("smtp.deshabilitado", "razon", "SMTP_PORT inválido", "valor", puertoStr)
			return
		}

		remitente := strings.TrimSpace(os.Getenv("SMTP_REMITENTE"))
		if remitente == "" {
			remitente = usuario
		}

		configCargada = ConfigSmtp{
			Host:       host,
			Puerto:     puerto,
			Usuario:    usuario,
			Password:   password,
			Remitente:  remitente,
			Habilitado: true,
		}
		slog.Info("smtp.cargado", "host", host, "puerto", puerto, "usuario", usuario)
	})
	return configCargada
}

// MensajeCorreo representa un email simple en texto plano. NO soportamos
// HTML, adjuntos, ni multipart — eso minimiza la chance de que el servidor
// rechace el mensaje por formato y nos cuente como error.
type MensajeCorreo struct {
	Destinatario string
	Asunto       string
	Cuerpo       string // texto plano, líneas <=78 chars idealmente
}

// Enviar manda el mensaje. Vuelve sin error si el SMTP no está configurado
// (no es excepción — el caller ya logueó el link como fallback). Si el
// circuit breaker está abierto, también vuelve sin error sin tocar la red.
func Enviar(contexto context.Context, msg MensajeCorreo) error {
	cfg := Configuracion()
	if !cfg.Habilitado {
		return errors.New("smtp no configurado")
	}
	if !cortocircuitoGlobal.PuedeIntentar() {
		return errors.New("smtp en pausa por errores consecutivos")
	}

	direccionDestino, err := mail.ParseAddress(msg.Destinatario)
	if err != nil {
		return fmt.Errorf("destinatario inválido: %w", err)
	}

	// Envío con timeout estricto. Si se pasa, abandonamos.
	tipoErr := make(chan error, 1)
	go func() { tipoErr <- enviarSmtps(cfg, direccionDestino.Address, msg) }()

	select {
	case err := <-tipoErr:
		if err != nil {
			pausado := cortocircuitoGlobal.RegistrarError()
			slog.Warn("smtp.envio_fallido",
				"destino", direccionDestino.Address,
				"error", err.Error(),
				"circuito_pausado", pausado,
			)
			return err
		}
		cortocircuitoGlobal.RegistrarExito()
		slog.Info("smtp.enviado", "destino", direccionDestino.Address)
		return nil

	case <-contexto.Done():
		// Timeout / cancelación del contexto HTTP: no contamos como error
		// del proveedor (no es culpa de él), pero abandonamos.
		return contexto.Err()

	case <-time.After(timeoutDial + timeoutEnvio):
		pausado := cortocircuitoGlobal.RegistrarError()
		slog.Warn("smtp.timeout_total", "destino", direccionDestino.Address, "circuito_pausado", pausado)
		return errors.New("timeout total enviando correo")
	}
}

// enviarSmtps abre la conversación SMTP eligiendo el modo según puerto:
//   - 465 → TLS implícito desde el TCP-dial (SMTPS clásico)
//   - cualquier otro (587 típicamente) → TCP plano + STARTTLS
//
// En el datacenter de producción el puerto 465 suele estar bloqueado por
// el ISP; 587 es el camino estándar para envíos autenticados modernos.
func enviarSmtps(cfg ConfigSmtp, destino string, msg MensajeCorreo) error {
	direccion := fmt.Sprintf("%s:%d", cfg.Host, cfg.Puerto)
	configTls := &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}

	cliente, err := abrirClienteSmtp(direccion, cfg, configTls)
	if err != nil {
		return err
	}
	defer cliente.Close()

	auth := smtp.PlainAuth("", cfg.Usuario, cfg.Password, cfg.Host)
	if err := cliente.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err := cliente.Mail(cfg.Usuario); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err := cliente.Rcpt(destino); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}

	escritor, err := cliente.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}

	encabezados := []string{
		"From: " + cfg.Remitente,
		"To: " + destino,
		"Subject: " + msg.Asunto,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
	}
	mensajeCompleto := strings.Join(encabezados, "\r\n") + msg.Cuerpo
	if _, err := escritor.Write([]byte(mensajeCompleto)); err != nil {
		return fmt.Errorf("escribir cuerpo: %w", err)
	}
	if err := escritor.Close(); err != nil {
		return fmt.Errorf("cerrar DATA: %w", err)
	}

	return cliente.Quit()
}

// abrirClienteSmtp encapsula la elección del modo de cifrado:
//
//   - puerto 465: TLS implícito. Hacemos tls.Dial y le pasamos la conexión
//     ya cifrada a smtp.NewClient. Funciona donde el ISP no bloquea 465.
//
//   - otros puertos (587, 25, etc.): TCP plano + STARTTLS upgrade. Es lo
//     que usan los relays modernos y lo que el datacenter típicamente
//     permite saliendo. Si el server no anuncia STARTTLS, abortamos —
//     no enviamos autenticado en claro NI bajo amenaza de bloqueo.
func abrirClienteSmtp(direccion string, cfg ConfigSmtp, configTls *tls.Config) (*smtp.Client, error) {
	dialer := &net.Dialer{Timeout: timeoutDial}

	if cfg.Puerto == 465 {
		conexion, err := tls.DialWithDialer(dialer, "tcp", direccion, configTls)
		if err != nil {
			return nil, fmt.Errorf("dial tls 465: %w", err)
		}
		_ = conexion.SetDeadline(time.Now().Add(timeoutEnvio))
		cliente, err := smtp.NewClient(conexion, cfg.Host)
		if err != nil {
			conexion.Close()
			return nil, fmt.Errorf("cliente smtp tls: %w", err)
		}
		return cliente, nil
	}

	conexion, err := dialer.Dial("tcp", direccion)
	if err != nil {
		return nil, fmt.Errorf("dial tcp %d: %w", cfg.Puerto, err)
	}
	_ = conexion.SetDeadline(time.Now().Add(timeoutEnvio))
	cliente, err := smtp.NewClient(conexion, cfg.Host)
	if err != nil {
		conexion.Close()
		return nil, fmt.Errorf("cliente smtp plano: %w", err)
	}
	soportaStartTls, _ := cliente.Extension("STARTTLS")
	if !soportaStartTls {
		cliente.Close()
		return nil, errors.New("servidor no anuncia STARTTLS y rechazamos enviar en claro")
	}
	if err := cliente.StartTLS(configTls); err != nil {
		cliente.Close()
		return nil, fmt.Errorf("STARTTLS: %w", err)
	}
	return cliente, nil
}
