package boveda

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yeka/zip"

	"sistemas-unificados/compartido/validaciones"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

// ItemExportable: lo que va dentro del JSON exportado. Plaintext dentro del
// archivo cifrado — el cifrado AES-256 del ZIP lo protege.
type ItemExportable struct {
	Titulo         string `json:"titulo"`
	SistemaCodigo  string `json:"sistema_codigo"`
	SistemaNombre  string `json:"sistema_nombre"`
	SistemaUrl     string `json:"sistema_url"`
	UsuarioExterno string `json:"usuario_externo"`
	Password       string `json:"password"`
	Observaciones  string `json:"observaciones,omitempty"`
	Tipo           string `json:"tipo"`
	Puerto         *int16 `json:"puerto,omitempty"`
	Estado         string `json:"estado"`
	CreadoEn       string `json:"creado_en"`
}

// BoletinExportacion: el JSON dentro del ZIP cifrado.
type BoletinExportacion struct {
	Version      int              `json:"version"`
	GeneradoEn   string           `json:"generado_en"`
	TotalAccesos int              `json:"total_accesos"`
	Accesos      []ItemExportable `json:"accesos"`
}

// ExportarBoveda genera un ZIP cifrado con AES-256 que contiene un único
// archivo JSON con TODOS los accesos descifrados.
//
// Seguridad:
//   - Passphrase obligatoria con fortaleza igual a passwords de operadores.
//   - Cifrado AES-256 (no "ZipCrypto" legacy).
//   - El JSON nunca toca disco — todo en memoria hasta enviar al cliente.
//
// El archivo resultante se abre con 7-Zip o WinRAR (Windows nativo NO
// soporta AES-256 ZIP; tema documentado en el LEEME que va adentro).
func ExportarBoveda(
	contexto context.Context,
	ejecutor cockroach.EjecutorSql,
	claves *cripto.ClavesCifrado,
	passphrase string,
) ([]byte, error) {
	if err := validaciones.ValidarFortalezaPassword(passphrase); err != nil {
		return nil, fmt.Errorf("passphrase débil: %w", err)
	}

	items, err := listarTodosParaExportar(contexto, ejecutor, claves)
	if err != nil {
		return nil, err
	}

	boletin := BoletinExportacion{
		Version:      1,
		GeneradoEn:   time.Now().UTC().Format(time.RFC3339),
		TotalAccesos: len(items),
		Accesos:      items,
	}
	cuerpoJson, err := json.MarshalIndent(boletin, "", "  ")
	if err != nil {
		return nil, err
	}

	bufZip := new(bytes.Buffer)
	escritorZip := zip.NewWriter(bufZip)

	// JSON CIFRADO: este es el archivo principal — único cifrado AES-256.
	nombreArchivo := fmt.Sprintf("gestor-codeplex-%s.json", time.Now().UTC().Format("2006-01-02"))
	escritorEntrada, err := escritorZip.Encrypt(nombreArchivo, passphrase, zip.AES256Encryption)
	if err != nil {
		return nil, err
	}
	if _, err := escritorEntrada.Write(cuerpoJson); err != nil {
		return nil, err
	}

	// LEEME público — sin cifrar para que el operador sepa cómo abrir el JSON.
	// No revela credenciales, solo instrucciones genéricas de descifrado.
	readme, err := escritorZip.Create("LEEME.txt")
	if err != nil {
		return nil, err
	}
	textoReadme := fmt.Sprintf(`Backup cifrado del Gestor de Accesos Codeplex
==============================================

Generado: %s
Accesos incluidos: %d

Este ZIP contiene un archivo JSON cifrado con AES-256. Para abrirlo
necesitas la passphrase que elegiste al exportar.

IMPORTANTE: Windows 11 nativo NO soporta AES-256 en ZIPs.
Usa una de estas herramientas:

OPCION 1 - 7-Zip (Windows/Mac/Linux, gratis, recomendado):
  Descargar: https://www.7-zip.org/
  Uso: click derecho sobre el .zip -> 7-Zip -> Open archive
       -> doble click al .json -> ingresar passphrase

OPCION 2 - WinRAR (Windows):
  Doble click al .zip -> doble click al .json -> ingresar passphrase

OPCION 3 - Linux/Mac línea de comandos:
  unzip -P 'tu_passphrase' gestor-codeplex-*.zip

Si pierdes la passphrase, los datos son IRRECUPERABLES.
Guarda este archivo y la passphrase en lugares separados:
  - El ZIP: USB, Drive, correo, donde quieras (sin la passphrase
    es indistinguible de ruido aleatorio)
  - La passphrase: tu password manager personal (1Password,
    Bitwarden, KeePass, etc.)
`, boletin.GeneradoEn, boletin.TotalAccesos)
	if _, err := readme.Write([]byte(textoReadme)); err != nil {
		return nil, err
	}

	if err := escritorZip.Close(); err != nil {
		return nil, err
	}

	slog.Info("boveda.exportada",
		"total_accesos", boletin.TotalAccesos,
		"bytes_zip", bufZip.Len(),
	)
	return bufZip.Bytes(), nil
}

// listarTodosParaExportar consulta y descifra todos los accesos vigentes
// (ACTIVO + REVOCADO). El password sí se incluye — es lo que hace que valga
// la pena el export. Por eso este path REQUIERE 2FA en el handler.
func listarTodosParaExportar(
	contexto context.Context,
	ejecutor cockroach.EjecutorSql,
	claves *cripto.ClavesCifrado,
) ([]ItemExportable, error) {
	filas, err := ejecutor.Query(contexto, `
		SELECT a.titulo,
		       coalesce(s.codigo, ''),
		       coalesce(s.nombre, ''),
		       coalesce(s.url_acceso, ''),
		       a.usuario_externo,
		       a.password_cifrada,
		       coalesce(a.observaciones, ''::BYTES),
		       a.tipo, a.puerto, a.estado, a.creado_en
		FROM acceso_guardado a
		LEFT JOIN sistema_destino s ON s.id = a.sistema_destino_id
		WHERE a.estado != 'ELIMINADO'
		ORDER BY a.creado_en ASC
	`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	clave := claves.ClaveBoveda()
	resultado := make([]ItemExportable, 0)
	omitidos := 0
	for filas.Next() {
		var item ItemExportable
		var usuarioCifrado, passwordCifrada, observacionesCifradas []byte
		var creadoEn time.Time
		if err := filas.Scan(
			&item.Titulo, &item.SistemaCodigo, &item.SistemaNombre, &item.SistemaUrl,
			&usuarioCifrado, &passwordCifrada, &observacionesCifradas,
			&item.Tipo, &item.Puerto, &item.Estado, &creadoEn,
		); err != nil {
			return nil, err
		}
		item.CreadoEn = creadoEn.UTC().Format(time.RFC3339)

		usuario, err := cripto.DescifrarConAesGcm(clave, usuarioCifrado)
		if err != nil {
			slog.Warn("exportar.fila_omitida", "razon", "no se pudo descifrar usuario", "error", err.Error())
			omitidos++
			continue
		}
		item.UsuarioExterno = string(usuario)

		password, err := cripto.DescifrarConAesGcm(clave, passwordCifrada)
		if err != nil {
			slog.Warn("exportar.fila_omitida", "razon", "no se pudo descifrar password", "error", err.Error())
			omitidos++
			continue
		}
		item.Password = string(password)

		if len(observacionesCifradas) > 0 {
			obs, err := cripto.DescifrarConAesGcm(clave, observacionesCifradas)
			if err == nil {
				item.Observaciones = string(obs)
			}
		}

		resultado = append(resultado, item)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}
	if omitidos > 0 {
		slog.Warn("exportar.resumen_omitidos", "filas_omitidas", omitidos)
	}
	return resultado, nil
}

// ErrPassphraseRequerida: handler la usa para distinguir error de validación.
var ErrPassphraseRequerida = errors.New("passphrase requerida y debe cumplir reglas de fortaleza")
