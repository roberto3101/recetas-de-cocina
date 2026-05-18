package boveda_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

const correoOperadorBoveda = "boveda-test@codeplex.pe"

func cargarEntornoOSaltar(t *testing.T) (*cockroach.ConexionBaseDatos, *cripto.ClavesCifrado) {
	t.Helper()
	url := os.Getenv("BASE_DATOS_URL")
	if url == "" {
		url = "postgresql://root@localhost:26258/sistemas_unificados?sslmode=disable"
	}
	if os.Getenv("KEK_GESTOR") == "" {
		k, err := cripto.GenerarClaveAleatoriaAES256()
		if err != nil {
			t.Fatalf("no se generó KEK: %v", err)
		}
		os.Setenv("KEK_GESTOR", k)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conexion, err := cockroach.AbrirConexion(ctx, url)
	if err != nil {
		t.Skipf("no se pudo conectar a CockroachDB local; test saltado: %v", err)
	}
	claves, err := cripto.CargarClavesDesdeEntorno()
	if err != nil {
		t.Fatalf("claves: %v", err)
	}
	return conexion, claves
}

func crearOperadorYSistema(t *testing.T, conexion *cockroach.ConexionBaseDatos, claves *cripto.ClavesCifrado) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// Limpia previo
	_, _ = conexion.Pool().Exec(ctx, `DELETE FROM acceso_guardado WHERE creado_por IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`, correoOperadorBoveda)
	_, _ = conexion.Pool().Exec(ctx, `DELETE FROM sistema_destino WHERE codigo='boveda_test_sys'`)
	_, _ = conexion.Pool().Exec(ctx, `DELETE FROM auditoria_accion WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`, correoOperadorBoveda)
	_, _ = conexion.Pool().Exec(ctx, `DELETE FROM sesion_global WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`, correoOperadorBoveda)
	_, _ = conexion.Pool().Exec(ctx, `DELETE FROM usuario_totp WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(correo_electronico) = lower($1))`, correoOperadorBoveda)
	_, _ = conexion.Pool().Exec(ctx, `DELETE FROM usuario WHERE lower(correo_electronico) = lower($1)`, correoOperadorBoveda)

	reg, err := identidad.RegistrarUsuarioYActivar(ctx, conexion, identidad.DatosRegistroUsuario{
		CorreoElectronico: correoOperadorBoveda,
		PasswordPlana:     "Operador_Test_2026!",
	})
	if err != nil {
		t.Fatalf("registrar operador: %v", err)
	}

	res, err := catalogo_sistemas.RegistrarSistema(ctx, conexion, claves, catalogo_sistemas.DatosRegistrarSistema{
		Codigo:    "boveda_test_sys",
		Nombre:    "Sistema test bóveda",
		UrlAcceso: "https://test.example.com/login",
		CreadoPor: reg.UsuarioId,
	})
	if err != nil {
		t.Fatalf("registrar sistema: %v", err)
	}

	return reg.UsuarioId, res.SistemaId
}

func TestBoveda_GuardarYConsultar_RoundtripCifrado(t *testing.T) {
	conexion, claves := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	operadorId, sistemaId := crearOperadorYSistema(t, conexion, claves)
	ctx := context.Background()

	acceso := &boveda.AccesoGuardado{
		Titulo:           "Mi acceso",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "juan@externo.com",
		PasswordPlana:    "ClaveSuperSecreta_2026!",
		Observaciones:    "Cuenta del jefe",
		CreadoPor:        &operadorId,
	}
	if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, acceso); err != nil {
		t.Fatalf("GuardarAcceso: %v", err)
	}
	if acceso.Id == uuid.Nil {
		t.Fatal("id no asignado tras guardar")
	}

	leido, err := boveda.ConsultarAccesoPorId(ctx, conexion.Pool(), claves, acceso.Id)
	if err != nil {
		t.Fatalf("ConsultarAccesoPorId: %v", err)
	}
	if leido.UsuarioExterno != "juan@externo.com" {
		t.Errorf("usuario descifrado distinto: %q", leido.UsuarioExterno)
	}
	if leido.PasswordPlana != "ClaveSuperSecreta_2026!" {
		t.Errorf("password descifrada distinta: %q", leido.PasswordPlana)
	}
	if leido.Observaciones != "Cuenta del jefe" {
		t.Errorf("observaciones descifradas distintas: %q", leido.Observaciones)
	}
}

func TestBoveda_BDEsCifrada_NoSeVeUsuarioEnPlano(t *testing.T) {
	conexion, claves := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	operadorId, sistemaId := crearOperadorYSistema(t, conexion, claves)
	ctx := context.Background()

	acceso := &boveda.AccesoGuardado{
		Titulo:           "Cifrado check",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "DEBE_NO_APARECER_EN_BD@example.com",
		PasswordPlana:    "CLAVE_DEBE_NO_APARECER_2026",
		Observaciones:    "OBS_DEBE_NO_APARECER",
		CreadoPor:        &operadorId,
	}
	if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, acceso); err != nil {
		t.Fatalf("guardar: %v", err)
	}

	var matches int64
	err := conexion.Pool().QueryRow(ctx, `
		SELECT count(*) FROM acceso_guardado
		WHERE position(convert_to($1, 'UTF8') IN usuario_externo) > 0
		   OR position(convert_to($2, 'UTF8') IN password_cifrada) > 0
		   OR position(convert_to($3, 'UTF8') IN observaciones) > 0
	`,
		"DEBE_NO_APARECER",
		"CLAVE_DEBE_NO_APARECER",
		"OBS_DEBE_NO_APARECER",
	).Scan(&matches)
	if err != nil {
		t.Fatalf("query verificación: %v", err)
	}
	if matches > 0 {
		t.Fatal("texto plano de usuario/password/observaciones aparece en BD; NO está cifrado correctamente")
	}
}

func TestBoveda_UniqueHashRechazaDuplicado(t *testing.T) {
	conexion, claves := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	operadorId, sistemaId := crearOperadorYSistema(t, conexion, claves)
	ctx := context.Background()

	primero := &boveda.AccesoGuardado{
		Titulo:           "Primero",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "duplicado@externo.com",
		PasswordPlana:    "ClaveA_2026!",
		CreadoPor:        &operadorId,
	}
	if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, primero); err != nil {
		t.Fatalf("primer guardado: %v", err)
	}

	// Segundo guardado con MISMO usuario_externo + sistema → ON CONFLICT debería actualizar (no fallar)
	segundo := &boveda.AccesoGuardado{
		Titulo:           "Segundo (debería actualizar el anterior)",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "duplicado@externo.com",
		PasswordPlana:    "ClaveB_2026!",
		CreadoPor:        &operadorId,
	}
	if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, segundo); err != nil {
		t.Fatalf("segundo guardado debería hacer upsert: %v", err)
	}

	leido, err := boveda.ConsultarAccesoPorId(ctx, conexion.Pool(), claves, segundo.Id)
	if err != nil {
		t.Fatalf("consulta tras upsert: %v", err)
	}
	if leido.Titulo != "Segundo (debería actualizar el anterior)" {
		t.Errorf("upsert no actualizó título: %q", leido.Titulo)
	}
	if leido.PasswordPlana != "ClaveB_2026!" {
		t.Errorf("upsert no actualizó password")
	}
}

func TestBoveda_DesactivarYReactivar(t *testing.T) {
	conexion, claves := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	operadorId, sistemaId := crearOperadorYSistema(t, conexion, claves)
	ctx := context.Background()

	acceso := &boveda.AccesoGuardado{
		Titulo:           "Para desactivar",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "desactivable@externo.com",
		PasswordPlana:    "Clave_2026!",
		CreadoPor:        &operadorId,
	}
	if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, acceso); err != nil {
		t.Fatalf("guardar: %v", err)
	}

	if err := boveda.DesactivarAcceso(ctx, conexion.Pool(), acceso.Id, operadorId); err != nil {
		t.Fatalf("desactivar: %v", err)
	}

	leido, err := boveda.ConsultarAccesoPorId(ctx, conexion.Pool(), claves, acceso.Id)
	if err != nil {
		t.Fatalf("consulta tras desactivar: %v", err)
	}
	if leido.Estado != "REVOCADO" {
		t.Errorf("estado tras desactivar esperado REVOCADO, got %s", leido.Estado)
	}

	if err := boveda.ReactivarAcceso(ctx, conexion.Pool(), acceso.Id, operadorId); err != nil {
		t.Fatalf("reactivar: %v", err)
	}
	leido, _ = boveda.ConsultarAccesoPorId(ctx, conexion.Pool(), claves, acceso.Id)
	if leido.Estado != "ACTIVO" {
		t.Errorf("estado tras reactivar esperado ACTIVO, got %s", leido.Estado)
	}
}

func TestBoveda_ActualizarSinNuevaPassword_ConservaLaAnterior(t *testing.T) {
	conexion, claves := cargarEntornoOSaltar(t)
	defer conexion.Cerrar()
	operadorId, sistemaId := crearOperadorYSistema(t, conexion, claves)
	ctx := context.Background()

	acceso := &boveda.AccesoGuardado{
		Titulo:           "Original",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "edit@externo.com",
		PasswordPlana:    "PwdOriginal_2026!",
		CreadoPor:        &operadorId,
	}
	if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, acceso); err != nil {
		t.Fatalf("guardar: %v", err)
	}

	editado := &boveda.AccesoGuardado{
		Id:               acceso.Id,
		Titulo:           "Cambió el título",
		SistemaDestinoId: sistemaId,
		UsuarioExterno:   "edit@externo.com",
		PasswordPlana:    "", // ← vacía: no toca password
		Observaciones:    "nuevas notas",
	}
	if err := boveda.ActualizarAcceso(ctx, conexion.Pool(), claves, editado, operadorId); err != nil {
		t.Fatalf("actualizar: %v", err)
	}

	leido, err := boveda.ConsultarAccesoPorId(ctx, conexion.Pool(), claves, acceso.Id)
	if err != nil {
		t.Fatalf("consulta tras update: %v", err)
	}
	if leido.Titulo != "Cambió el título" {
		t.Errorf("título no se actualizó: %q", leido.Titulo)
	}
	if leido.PasswordPlana != "PwdOriginal_2026!" {
		t.Errorf("password debería conservarse cuando PasswordPlana vacía; got %q", leido.PasswordPlana)
	}
	if leido.Observaciones != "nuevas notas" {
		t.Errorf("observaciones no se actualizaron: %q", leido.Observaciones)
	}
}
