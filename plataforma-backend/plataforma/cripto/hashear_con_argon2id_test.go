package cripto

import (
	"strings"
	"testing"
)

func TestHashearYVerificarArgon2id_RoundTrip(t *testing.T) {
	passwordOriginal := "P@ssw0rd_Seguro_2026"

	hash, err := HashearPasswordConArgon2id(passwordOriginal)
	if err != nil {
		t.Fatalf("HashearPasswordConArgon2id falló: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("hash no tiene prefijo argon2id: %s", hash)
	}

	coincide, err := VerificarPasswordContraHashArgon2id(passwordOriginal, hash)
	if err != nil {
		t.Fatalf("verificación falló: %v", err)
	}
	if !coincide {
		t.Fatal("la contraseña correcta no coincidió con su hash")
	}
}

func TestVerificarPasswordIncorrecta_NoCoincide(t *testing.T) {
	hash, err := HashearPasswordConArgon2id("password_correcta_aaa1!")
	if err != nil {
		t.Fatalf("hashear falló: %v", err)
	}
	coincide, err := VerificarPasswordContraHashArgon2id("password_incorrecta_xxx", hash)
	if err != nil {
		t.Fatalf("verificación falló: %v", err)
	}
	if coincide {
		t.Fatal("password incorrecta no debería coincidir")
	}
}

func TestDosHashesDeLaMismaPassword_DebenDiferir(t *testing.T) {
	password := "TheSamePassword#1"
	h1, err := HashearPasswordConArgon2id(password)
	if err != nil {
		t.Fatalf("h1 falló: %v", err)
	}
	h2, err := HashearPasswordConArgon2id(password)
	if err != nil {
		t.Fatalf("h2 falló: %v", err)
	}
	if h1 == h2 {
		t.Fatal("dos hashes de la misma password deberían diferir por la sal aleatoria")
	}
}

func TestVerificarHashSintaxisInvalida_DevuelveError(t *testing.T) {
	if _, err := VerificarPasswordContraHashArgon2id("x", "$bcrypt$v=10$xxx"); err == nil {
		t.Fatal("verificar hash con prefijo no argon2id debería retornar error")
	}
	if _, err := VerificarPasswordContraHashArgon2id("x", "esto no es un hash"); err == nil {
		t.Fatal("verificar texto sin estructura argon2 debería retornar error")
	}
}
