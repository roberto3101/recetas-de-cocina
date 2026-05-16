package integracion

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

const correoOperadorPruebas = "integracion@codeplex.pe"

func cargarEntornoOSaltar(t *testing.T) *cockroach.ConexionBaseDatos {
	t.Helper()
	urlConexion := os.Getenv("BASE_DATOS_URL")
	if urlConexion == "" {
		urlConexion = "postgresql://root@localhost:26258/sistemas_unificados?sslmode=disable"
	}
	if os.Getenv("KEK_GESTOR") == "" {
		clave, err := cripto.GenerarClaveAleatoriaAES256()
		if err != nil {
			t.Fatalf("no se generó KEK de prueba: %v", err)
		}
		os.Setenv("KEK_GESTOR", clave)
	}
	contexto, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	conexion, err := cockroach.AbrirConexion(contexto, urlConexion)
	if err != nil {
		t.Skipf("no se pudo conectar a BD local (CockroachDB en localhost:26258); test saltado: %v", err)
	}
	if _, err := cripto.CargarClavesDesdeEntorno(); err != nil {
		t.Fatalf("no se cargaron claves: %v", err)
	}
	return conexion
}

func limpiarOperadorDePrueba(t *testing.T, conexion *cockroach.ConexionBaseDatos) {
	t.Helper()
	contexto := context.Background()
	_, _ = conexion.Pool().Exec(contexto, `
		DELETE FROM auditoria_accion WHERE usuario_id IN (
			SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1)
		)
	`, correoOperadorPruebas)
	_, _ = conexion.Pool().Exec(contexto, `
		DELETE FROM sesion_global WHERE usuario_id IN (
			SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1)
		)
	`, correoOperadorPruebas)
	_, _ = conexion.Pool().Exec(contexto, `
		DELETE FROM usuario_totp WHERE usuario_id IN (
			SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1)
		)
	`, correoOperadorPruebas)
	_, _ = conexion.Pool().Exec(contexto, `
		DELETE FROM usuario WHERE lower(correo_electronico) = lower($1)
	`, correoOperadorPruebas)
}

func TestFlujoCompleto_RegistrarLoginYCerrarSesion(t *testing.T) {
	conexion := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	defer limpiarOperadorDePrueba(t, conexion)
	limpiarOperadorDePrueba(t, conexion)

	contexto := context.Background()

	resultadoReg, err := identidad.RegistrarUsuarioYActivar(contexto, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "P@ssw0rd_Operador_2026",
		IpOrigen:          "127.0.0.1",
		AgenteUsuario:     "test-suite",
	})
	if err != nil {
		t.Fatalf("RegistrarUsuarioYActivar falló: %v", err)
	}
	if resultadoReg.Estado != identidad.EstadoUsuarioActivo {
		t.Fatalf("estado esperado ACTIVO, got %s", resultadoReg.Estado)
	}

	resultadoSesion, err := identidad.IniciarSesion(contexto, conexion, identidad.DatosIniciarSesion{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "P@ssw0rd_Operador_2026",
		IpOrigen:          "127.0.0.1",
		AgenteUsuario:     "test-suite",
	})
	if err != nil {
		t.Fatalf("IniciarSesion falló: %v", err)
	}
	if resultadoSesion.TokenPlano == "" {
		t.Fatal("token plano vacío")
	}

	sesionActiva, err := identidad.ValidarToken(contexto, conexion, resultadoSesion.TokenPlano)
	if err != nil {
		t.Fatalf("ValidarToken falló: %v", err)
	}
	if sesionActiva.UsuarioId != resultadoReg.UsuarioId {
		t.Fatal("validación devolvió usuario distinto")
	}

	if err := identidad.CerrarSesion(contexto, conexion, identidad.DatosCerrarSesion{
		SesionId:      resultadoSesion.SesionId,
		UsuarioId:     resultadoReg.UsuarioId,
		IpOrigen:      "127.0.0.1",
		AgenteUsuario: "test-suite",
	}); err != nil {
		t.Fatalf("CerrarSesion falló: %v", err)
	}

	if _, err := identidad.ValidarToken(contexto, conexion, resultadoSesion.TokenPlano); !errors.Is(err, identidad.ErrSesionRevocada) {
		t.Fatalf("token revocado debería retornar ErrSesionRevocada, got: %v", err)
	}
}

func TestRegistrarUsuario_RechazaCorreoDuplicado(t *testing.T) {
	conexion := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	defer limpiarOperadorDePrueba(t, conexion)
	limpiarOperadorDePrueba(t, conexion)

	contexto := context.Background()

	_, err := identidad.RegistrarUsuarioYActivar(contexto, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "P@ssw0rd_2026_test",
	})
	if err != nil {
		t.Fatalf("primer registro falló: %v", err)
	}

	_, err = identidad.RegistrarUsuario(contexto, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "Otra_Password_2026!",
	})
	if !errors.Is(err, identidad.ErrCorreoYaRegistrado) {
		t.Fatalf("se esperaba ErrCorreoYaRegistrado, got: %v", err)
	}
}

func TestIniciarSesion_RechazaPasswordIncorrecto(t *testing.T) {
	conexion := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	defer limpiarOperadorDePrueba(t, conexion)
	limpiarOperadorDePrueba(t, conexion)

	contexto := context.Background()

	if _, err := identidad.RegistrarUsuarioYActivar(contexto, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "Password_Real_2026!",
	}); err != nil {
		t.Fatalf("registro falló: %v", err)
	}

	_, err := identidad.IniciarSesion(contexto, conexion, identidad.DatosIniciarSesion{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "password_falso",
	})
	if !errors.Is(err, identidad.ErrCredencialesInvalidas) {
		t.Fatalf("se esperaba ErrCredencialesInvalidas, got: %v", err)
	}
}

func TestCambiarPassword_InvalidaSesionesActivas(t *testing.T) {
	conexion := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	defer limpiarOperadorDePrueba(t, conexion)
	limpiarOperadorDePrueba(t, conexion)

	contexto := context.Background()

	reg, err := identidad.RegistrarUsuarioYActivar(contexto, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "Password_Inicial_2026!",
	})
	if err != nil {
		t.Fatal(err)
	}

	sesion, err := identidad.IniciarSesion(contexto, conexion, identidad.DatosIniciarSesion{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "Password_Inicial_2026!",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := identidad.CambiarPassword(contexto, conexion, identidad.DatosCambioPassword{
		UsuarioId:      reg.UsuarioId,
		PasswordActual: "Password_Inicial_2026!",
		PasswordNueva:  "Password_NUEVO_2026!",
	}); err != nil {
		t.Fatalf("CambiarPassword falló: %v", err)
	}

	if _, err := identidad.ValidarToken(contexto, conexion, sesion.TokenPlano); err == nil {
		t.Fatal("la sesión previa debería estar invalidada tras cambio de password")
	}

	if _, err := identidad.IniciarSesion(contexto, conexion, identidad.DatosIniciarSesion{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "Password_NUEVO_2026!",
	}); err != nil {
		t.Fatalf("login con password nuevo falló: %v", err)
	}

	if _, err := identidad.IniciarSesion(contexto, conexion, identidad.DatosIniciarSesion{
		CorreoElectronico: correoOperadorPruebas,
		PasswordPlana:     "Password_Inicial_2026!",
	}); !errors.Is(err, identidad.ErrCredencialesInvalidas) {
		t.Fatalf("login con password viejo debería fallar, got: %v", err)
	}
}
