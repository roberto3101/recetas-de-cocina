package main

import (
	"fmt"
	"os"

	"sistemas-unificados/plataforma/cripto"
)

// Herramienta one-shot: genera un hash Argon2id usando los mismos parametros
// que el resto del backend (cripto.HashearPasswordConArgon2id). Solo para
// crear usuarios manualmente cuando no hay sesion con 2FA disponible.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: generar_hash <password_plana>")
		os.Exit(1)
	}
	hash, err := cripto.HashearPasswordConArgon2id(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(hash)
}
