package cripto

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerarSecretoTotp_DevuelveSecretoValido(t *testing.T) {
	clave, err := GenerarSecretoTotp("GestorPruebas", "ana@codeplex.pe")
	if err != nil {
		t.Fatalf("GenerarSecretoTotp falló: %v", err)
	}
	if clave.Secret() == "" {
		t.Fatal("secreto vacío")
	}
	if !strings.Contains(clave.URL(), "otpauth://totp/") {
		t.Fatalf("URL no tiene prefijo otpauth: %s", clave.URL())
	}
}

func TestValidarCodigoTotp_AceptaCodigoValido(t *testing.T) {
	clave, err := GenerarSecretoTotp("GestorPruebas", "luis@codeplex.pe")
	if err != nil {
		t.Fatalf("falló: %v", err)
	}
	codigoActual, err := totp.GenerateCode(clave.Secret(), time.Now())
	if err != nil {
		t.Fatalf("no se pudo generar código: %v", err)
	}
	if !ValidarCodigoTotp(codigoActual, clave.Secret()) {
		t.Fatal("código TOTP recién generado debería ser válido")
	}
}

func TestValidarCodigoTotp_RechazaCodigoInventado(t *testing.T) {
	clave, err := GenerarSecretoTotp("GestorPruebas", "x@y.com")
	if err != nil {
		t.Fatalf("falló: %v", err)
	}
	if ValidarCodigoTotp("000000", clave.Secret()) {
		t.Fatal("código fijo 000000 no debería validar para secreto aleatorio")
	}
}

func TestGenerarCodigosRespaldo_DevuelveDiezCodigosUnicos(t *testing.T) {
	planos, hashes, err := GenerarCodigosRespaldoTotp()
	if err != nil {
		t.Fatalf("falló: %v", err)
	}
	if len(planos) != 10 || len(hashes) != 10 {
		t.Fatalf("se esperaban 10 códigos y 10 hashes, got %d/%d", len(planos), len(hashes))
	}
	vistos := make(map[string]bool)
	for _, p := range planos {
		if vistos[p] {
			t.Fatalf("código respaldo repetido: %s", p)
		}
		vistos[p] = true
	}
	for i, hash := range hashes {
		if hash != HashearTokenConSha256(planos[i]) {
			t.Fatalf("hash %d no corresponde a plano", i)
		}
	}
}
