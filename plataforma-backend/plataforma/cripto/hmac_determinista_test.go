package cripto

import (
	"bytes"
	"testing"
)

func TestHmacSha256_DevuelveMismoHashParaMismoInput(t *testing.T) {
	clave := []byte("clave-de-prueba-32-bytes-1234567")

	h1 := HmacSha256(clave, "usuario@ejemplo.com")
	h2 := HmacSha256(clave, "usuario@ejemplo.com")

	if !bytes.Equal(h1, h2) {
		t.Fatalf("HMAC determinístico debería retornar mismo valor para mismo input; h1=%x h2=%x", h1, h2)
	}
	if len(h1) != 32 {
		t.Fatalf("HMAC-SHA256 debería ser 32 bytes; got %d", len(h1))
	}
}

func TestHmacSha256_HashDiferenteParaInputsDiferentes(t *testing.T) {
	clave := []byte("clave-de-prueba-32-bytes-1234567")

	h1 := HmacSha256(clave, "usuario@ejemplo.com")
	h2 := HmacSha256(clave, "otro@ejemplo.com")

	if bytes.Equal(h1, h2) {
		t.Fatal("HMAC debería ser distinto para correos distintos")
	}
}

func TestHmacSha256_HashDiferenteParaClavesDiferentes(t *testing.T) {
	mensaje := "usuario@ejemplo.com"

	h1 := HmacSha256([]byte("clave-A-32-bytes-1234567890abcde"), mensaje)
	h2 := HmacSha256([]byte("clave-B-32-bytes-1234567890abcde"), mensaje)

	if bytes.Equal(h1, h2) {
		t.Fatal("HMAC con KEK distintos debería ser distinto, aunque el mensaje sea el mismo")
	}
}

func TestHmacSha256_NormalizaCorreoMayusculasYEspacios(t *testing.T) {
	clave := []byte("clave-de-prueba-32-bytes-1234567")

	h1 := HmacSha256(clave, "Usuario@Ejemplo.com")
	h2 := HmacSha256(clave, "usuario@ejemplo.com")
	h3 := HmacSha256(clave, "  usuario@ejemplo.com  ")

	if !bytes.Equal(h1, h2) {
		t.Fatal("HMAC debería ignorar mayúsculas (normaliza a lower)")
	}
	if !bytes.Equal(h2, h3) {
		t.Fatal("HMAC debería ignorar espacios al inicio/fin (normaliza con TrimSpace)")
	}
}

func TestHmacSha256_MensajeVacioProduceHashValido(t *testing.T) {
	clave := []byte("clave-de-prueba-32-bytes-1234567")
	h := HmacSha256(clave, "")
	if len(h) != 32 {
		t.Fatalf("aun con mensaje vacío debe retornar 32 bytes; got %d", len(h))
	}
}
