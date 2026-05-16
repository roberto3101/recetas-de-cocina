package cripto

import "testing"

func TestGenerarTokenSesion_HashCoincide(t *testing.T) {
	plano, hash, err := GenerarTokenSesionAleatorio()
	if err != nil {
		t.Fatalf("falló: %v", err)
	}
	if plano == "" {
		t.Fatal("token plano vacío")
	}
	if hash != HashearTokenConSha256(plano) {
		t.Fatal("hash devuelto no coincide con hash recalculado")
	}
}

func TestDosTokensConsecutivos_DebenDiferir(t *testing.T) {
	plano1, _, _ := GenerarTokenSesionAleatorio()
	plano2, _, _ := GenerarTokenSesionAleatorio()
	if plano1 == plano2 {
		t.Fatal("dos tokens aleatorios consecutivos no deberían coincidir")
	}
}

func TestHashearTokenSha256_EsDeterminista(t *testing.T) {
	plano := "token-de-prueba-12345"
	h1 := HashearTokenConSha256(plano)
	h2 := HashearTokenConSha256(plano)
	if h1 != h2 {
		t.Fatal("hash SHA-256 debe ser determinista")
	}
	if len(h1) != 64 {
		t.Fatalf("hash SHA-256 hex debe tener 64 caracteres, got %d", len(h1))
	}
}
