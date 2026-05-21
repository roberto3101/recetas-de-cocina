package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"sistemas-unificados/capacidades/boveda"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
)

// Seeder de prueba: crea CANTIDAD accesos marcados con titulo "[TEST] Acceso NNN"
// todos dentro de un sistema marcador "test_paginacion". Pensado para validar
// la paginación / búsqueda / orden a volumen real.
//
// Limpieza después del test, UNA línea (vía cockroach sql):
//
//   DELETE FROM acceso_guardado WHERE titulo LIKE '[TEST] Acceso %';
//   DELETE FROM sistema_destino  WHERE codigo = 'test_paginacion';
//
// Importante: el seeder es idempotente para el sistema_destino (lo reutiliza
// si ya existe), pero NO para los accesos. Si lo corres dos veces, el
// segundo run falla con ErrAccesoYaExiste por el UNIQUE INDEX sobre
// (sistema_destino_id, usuario_externo_hash). Borra primero con el SQL de
// arriba si querés volver a sembrar.

const (
	codigoSistemaTest = "test_paginacion"
	nombreSistemaTest = "[TEST] Paginación"
	urlTest           = "https://test.codeplex.pe/"
	cantidadDefault   = 200
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("info: no hay .env local, usando vars de entorno del proceso")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cantidad := cantidadDefault
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &cantidad)
		if cantidad <= 0 || cantidad > 5000 {
			log.Fatalf("cantidad fuera de rango (1..5000): %d", cantidad)
		}
	}

	claves, err := cripto.CargarClavesDesdeEntorno()
	if err != nil {
		log.Fatalf("claves: %v", err)
	}

	conexion, err := cockroach.AbrirConexion(ctx, os.Getenv("BASE_DATOS_URL"))
	if err != nil {
		log.Fatalf("conexion BD: %v", err)
	}
	defer conexion.Cerrar()

	sistemaId, err := obtenerOCrearSistemaTest(ctx, conexion)
	if err != nil {
		log.Fatalf("sistema_destino marcador: %v", err)
	}
	fmt.Printf("Sistema marcador listo: id=%s codigo=%s\n\n", sistemaId, codigoSistemaTest)

	fmt.Printf("Sembrando %d accesos de prueba...\n", cantidad)
	creados := 0
	for i := 1; i <= cantidad; i++ {
		a := &boveda.AccesoGuardado{
			Titulo:           fmt.Sprintf("[TEST] Acceso %04d", i),
			SistemaDestinoId: sistemaId,
			UsuarioExterno:   fmt.Sprintf("test.user.%04d@codeplex.pe", i),
			PasswordPlana:    fmt.Sprintf("ClaveTestSegura_%04d!", i),
			Observaciones:    fmt.Sprintf("Observación dummy del acceso %d para validar paginación y búsqueda", i),
			Tipo:             "WEB",
		}
		if err := boveda.GuardarAcceso(ctx, conexion.Pool(), claves, a); err != nil {
			log.Printf("  [%04d] error: %v", i, err)
			continue
		}
		creados++
		if i%20 == 0 {
			fmt.Printf("  ...%d/%d\n", i, cantidad)
		}
	}

	fmt.Printf("\nTotal creados: %d/%d\n\n", creados, cantidad)
	fmt.Printf("Para borrar TODO lo de este seeder en un solo paso:\n\n")
	fmt.Printf("  cockroach sql --insecure --host=localhost:26257 --database=sistemas_unificados <<'SQL'\n")
	fmt.Printf("    DELETE FROM acceso_guardado WHERE titulo LIKE '[TEST] Acceso %%';\n")
	fmt.Printf("    DELETE FROM sistema_destino  WHERE codigo = '%s';\n", codigoSistemaTest)
	fmt.Printf("  SQL\n\n")
}

// obtenerOCrearSistemaTest devuelve el id del sistema marcador. Si ya
// existe (re-ejecución), lo reutiliza. Si no, lo crea con un INSERT directo
// para evitar dependencias en handlers HTTP/sesión.
func obtenerOCrearSistemaTest(ctx context.Context, conexion *cockroach.ConexionBaseDatos) (uuid.UUID, error) {
	var id uuid.UUID
	err := conexion.Pool().QueryRow(ctx,
		`SELECT id FROM sistema_destino WHERE codigo = $1`,
		codigoSistemaTest,
	).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Insertar directamente. estado='ACTIVO' para que aparezca en /cocina/sistemas.
	err = conexion.Pool().QueryRow(ctx, `
		INSERT INTO sistema_destino (
			codigo, nombre, url_acceso, url_login,
			nombre_campo_usuario, nombre_campo_password, metodo_login, estado
		) VALUES ($1, $2, $3, $4, 'usuario', 'password', 'POST', 'ACTIVO')
		RETURNING id
	`,
		codigoSistemaTest, nombreSistemaTest, urlTest, urlTest,
	).Scan(&id)
	return id, err
}
