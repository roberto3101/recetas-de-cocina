package consumo_usuarios

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/plataforma/cripto"
)

func AbrirPoolExterno(contexto context.Context, conexion *catalogo_sistemas.ConexionLectura, clavesCifrado *cripto.ClavesCifrado) (*pgxpool.Pool, error) {
	clave := clavesCifrado.ClaveBoveda()
	usuarioPlano, err := cripto.DescifrarConAesGcm(clave, conexion.UsuarioDbCifrado)
	if err != nil {
		return nil, fmt.Errorf("no se pudo descifrar el usuario de la conexión: %w", err)
	}
	passwordPlana, err := cripto.DescifrarConAesGcm(clave, conexion.PasswordDbCifrada)
	if err != nil {
		return nil, fmt.Errorf("no se pudo descifrar la contraseña de la conexión: %w", err)
	}

	urlConexion := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(string(usuarioPlano)),
		url.QueryEscape(string(passwordPlana)),
		conexion.Host,
		conexion.Puerto,
		url.PathEscape(conexion.BaseDatos),
		conexion.SslModo,
	)

	configuracion, err := pgxpool.ParseConfig(urlConexion)
	if err != nil {
		return nil, err
	}
	configuracion.MaxConns = 5
	configuracion.MinConns = 1
	configuracion.MaxConnLifetime = 15 * time.Minute
	configuracion.MaxConnIdleTime = 3 * time.Minute

	pool, err := pgxpool.NewWithConfig(contexto, configuracion)
	if err != nil {
		return nil, err
	}

	contextoPing, cancelar := context.WithTimeout(contexto, 5*time.Second)
	defer cancelar()
	if err := pool.Ping(contextoPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping a base externa falló: %w", err)
	}

	return pool, nil
}
