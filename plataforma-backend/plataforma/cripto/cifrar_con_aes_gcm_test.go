package cripto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func generarClaveAleatoriaParaPrueba(t *testing.T) []byte {
	t.Helper()
	clave := make([]byte, 32)
	if _, err := rand.Read(clave); err != nil {
		t.Fatalf("no se pudo generar clave de prueba: %v", err)
	}
	return clave
}

func TestCifrarYDescifrarConAesGcm_RoundTrip(t *testing.T) {
	clave := generarClaveAleatoriaParaPrueba(t)
	textoOriginal := []byte("contraseña con tildes y carácteres ñ 🔒 1234")

	cifrado, err := CifrarConAesGcm(clave, textoOriginal)
	if err != nil {
		t.Fatalf("CifrarConAesGcm falló: %v", err)
	}
	if bytes.Equal(cifrado, textoOriginal) {
		t.Fatal("el texto cifrado no debería ser igual al plano")
	}

	plano, err := DescifrarConAesGcm(clave, cifrado)
	if err != nil {
		t.Fatalf("DescifrarConAesGcm falló: %v", err)
	}
	if !bytes.Equal(plano, textoOriginal) {
		t.Fatalf("round-trip no preserva el texto: got %q, want %q", plano, textoOriginal)
	}
}

func TestDescifrarConClaveIncorrecta_DebeFallar(t *testing.T) {
	clave1 := generarClaveAleatoriaParaPrueba(t)
	clave2 := generarClaveAleatoriaParaPrueba(t)

	cifrado, err := CifrarConAesGcm(clave1, []byte("secreto"))
	if err != nil {
		t.Fatalf("cifrado falló: %v", err)
	}

	if _, err := DescifrarConAesGcm(clave2, cifrado); err == nil {
		t.Fatal("descifrado con clave incorrecta debería fallar")
	}
}

func TestDescifrarConPaqueteCorrompido_DebeFallar(t *testing.T) {
	clave := generarClaveAleatoriaParaPrueba(t)
	cifrado, err := CifrarConAesGcm(clave, []byte("hola"))
	if err != nil {
		t.Fatalf("cifrado falló: %v", err)
	}
	cifrado[len(cifrado)-1] ^= 0x01

	if _, err := DescifrarConAesGcm(clave, cifrado); err == nil {
		t.Fatal("descifrado de paquete manipulado debería fallar")
	}
}

func TestDosCifradosDelMismoTexto_DebenSerDistintos(t *testing.T) {
	clave := generarClaveAleatoriaParaPrueba(t)
	cifrado1, err := CifrarConAesGcm(clave, []byte("idéntico"))
	if err != nil {
		t.Fatalf("cifrado 1 falló: %v", err)
	}
	cifrado2, err := CifrarConAesGcm(clave, []byte("idéntico"))
	if err != nil {
		t.Fatalf("cifrado 2 falló: %v", err)
	}
	if bytes.Equal(cifrado1, cifrado2) {
		t.Fatal("dos cifrados del mismo texto deberían diferir por el nonce aleatorio")
	}
}
