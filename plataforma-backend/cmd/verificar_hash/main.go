package main

import (
	"fmt"
	"os"

	"sistemas-unificados/plataforma/cripto"
)

// Herramienta one-shot: verifica si una password en texto plano coincide con
// un hash Argon2id. Util para validar que un usuario recién creado tiene la
// clave esperada antes de entregarsela al operador.
func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "uso: verificar_hash <password_plana> <hash>")
		os.Exit(1)
	}
	ok, err := cripto.VerificarPasswordContraHashArgon2id(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if ok {
		fmt.Println("COINCIDE")
		return
	}
	fmt.Println("NO_COINCIDE")
	os.Exit(2)
}
