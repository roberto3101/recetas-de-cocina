package http

import (
	"fmt"
	"net/http"
	"time"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
	codigosError "sistemas-unificados/plataforma/gobierno/errores"
)

type cuerpoExportar struct {
	Passphrase string `json:"passphrase"`
}

// ConstruirHandlerExportarBoveda devuelve un ZIP cifrado AES-256 con todos
// los accesos descifrados. Requiere 2FA (lo aplica el middleware del grupo
// con2FA en rutas.go).
//
// Seguridad:
//   - Passphrase obligatoria (validada con misma fortaleza que passwords).
//   - Audit log ANTES del envío: si por alguna razón el cliente desconecta
//     mid-descarga, igual queda registro de que se intentó.
//   - El ZIP cifrado puede ir por cualquier canal (correo, Drive) sin
//     riesgo - sin passphrase es indistinguible de ruido.
func ConstruirHandlerExportarBoveda(conexion *cockroach.ConexionBaseDatos, clavesCifrado *cripto.ClavesCifrado) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		sesion, ok := ObtenerSesionDelContexto(escritor, peticion)
		if !ok {
			return
		}

		var cuerpo cuerpoExportar
		if !LeerCuerpoJson(escritor, peticion, &cuerpo) {
			return
		}

		// Audit ANTES de generar el ZIP. Si el cliente se desconecta a
		// mitad del download, igual queda registro de la intención.
		// auditoria.RegistrarAccion requiere una transacción (pgx.Tx), no el pool.
		trans, err := conexion.Pool().Begin(peticion.Context())
		if err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		if err := auditoria.RegistrarAccion(peticion.Context(), trans, auditoria.EntradaAuditoria{
			UsuarioId:     &sesion.UsuarioId,
			SesionId:      &sesion.SesionId,
			Modulo:        "BOVEDA",
			Accion:        "BOVEDA_EXPORTADA",
			IpOrigen:      obtenerIpRemota(peticion),
			AgenteUsuario: peticion.UserAgent(),
		}); err != nil {
			_ = trans.Rollback(peticion.Context())
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}
		if err := trans.Commit(peticion.Context()); err != nil {
			ResponderError(escritor, http.StatusInternalServerError, codigosError.CodigoErrorInterno, err.Error())
			return
		}

		bytesZip, err := boveda.ExportarBoveda(peticion.Context(), conexion.Pool(), clavesCifrado, cuerpo.Passphrase)
		if err != nil {
			ResponderError(escritor, http.StatusBadRequest, codigosError.CodigoValidacionFallida, err.Error())
			return
		}

		nombreArchivo := fmt.Sprintf("gestor-codeplex-%s.zip", time.Now().UTC().Format("2006-01-02-150405"))
		escritor.Header().Set("Content-Type", "application/zip")
		escritor.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, nombreArchivo))
		escritor.Header().Set("Content-Length", fmt.Sprintf("%d", len(bytesZip)))
		escritor.WriteHeader(http.StatusOK)
		_, _ = escritor.Write(bytesZip)
	}
}
