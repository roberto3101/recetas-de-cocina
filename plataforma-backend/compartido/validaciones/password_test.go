package validaciones

import (
	"errors"
	"testing"
)

func TestValidarFortalezaPassword_AceptaFuerte(t *testing.T) {
	if err := ValidarFortalezaPassword("MiContra$3na2026"); err != nil {
		t.Fatalf("debería ser válida: %v", err)
	}
}

func TestValidarFortalezaPassword_RechazaCorta(t *testing.T) {
	if err := ValidarFortalezaPassword("Aa1!"); !errors.Is(err, ErrPasswordMuyCorto) {
		t.Fatalf("se esperaba ErrPasswordMuyCorto, got: %v", err)
	}
}

func TestValidarFortalezaPassword_RechazaSinMayuscula(t *testing.T) {
	if err := ValidarFortalezaPassword("password_largo_1!"); !errors.Is(err, ErrPasswordSinMayuscula) {
		t.Fatalf("se esperaba ErrPasswordSinMayuscula, got: %v", err)
	}
}

func TestValidarFortalezaPassword_RechazaSinEspecial(t *testing.T) {
	if err := ValidarFortalezaPassword("PasswordLargo123A"); !errors.Is(err, ErrPasswordSinEspecial) {
		t.Fatalf("se esperaba ErrPasswordSinEspecial, got: %v", err)
	}
}

func TestEsCorreoElectronicoValido(t *testing.T) {
	casos := map[string]bool{
		"juan@codeplex.pe":       true,
		"a@b.co":                 true,
		"sin-arroba":             false,
		"":                       false,
		"con espacio@codeplex.pe": false,
	}
	for entrada, esperado := range casos {
		if EsCorreoElectronicoValido(entrada) != esperado {
			t.Errorf("EsCorreoElectronicoValido(%q) = %v; want %v", entrada, !esperado, esperado)
		}
	}
}
