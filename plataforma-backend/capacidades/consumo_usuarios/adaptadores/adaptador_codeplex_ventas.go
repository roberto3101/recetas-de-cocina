package adaptadores

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sistemas-unificados/plataforma/cripto"
)

const ClaveCodeplexVentas = "codeplex_ventas"

type AdaptadorCodeplexVentas struct {
	pool *pgxpool.Pool
}

func NuevoAdaptadorCodeplexVentas(pool *pgxpool.Pool) *AdaptadorCodeplexVentas {
	return &AdaptadorCodeplexVentas{pool: pool}
}

func (a *AdaptadorCodeplexVentas) Clave() string {
	return ClaveCodeplexVentas
}

func (a *AdaptadorCodeplexVentas) ProbarConexion(contexto context.Context) error {
	return a.pool.Ping(contexto)
}

func (a *AdaptadorCodeplexVentas) ListarUsuarios(contexto context.Context, filtro FiltroBusquedaUsuarios) (*ResultadoBusquedaUsuarios, error) {
	if filtro.TamanoPagina <= 0 {
		filtro.TamanoPagina = 100
	}
	if filtro.TamanoPagina > 200 {
		filtro.TamanoPagina = 200
	}
	if filtro.Pagina < 1 {
		filtro.Pagina = 1
	}
	desplazamiento := (filtro.Pagina - 1) * filtro.TamanoPagina

	var condiciones []string
	var argumentos []any
	indice := 1

	if filtro.Texto != "" {
		condiciones = append(condiciones, fmt.Sprintf("lower(correo_electronico) LIKE $%d", indice))
		argumentos = append(argumentos, "%"+strings.ToLower(strings.TrimSpace(filtro.Texto))+"%")
		indice++
	}
	if filtro.Estado != "" {
		condiciones = append(condiciones, fmt.Sprintf("estado = $%d", indice))
		argumentos = append(argumentos, filtro.Estado)
		indice++
	}

	donde := ""
	if len(condiciones) > 0 {
		donde = "WHERE " + strings.Join(condiciones, " AND ")
	}

	var total int64
	if err := a.pool.QueryRow(contexto,
		"SELECT count(*) FROM usuario "+donde, argumentos...,
	).Scan(&total); err != nil {
		return nil, err
	}

	consulta := `
		SELECT id::TEXT, correo_electronico, coalesce(password_hash,''), estado
		FROM usuario ` + donde + fmt.Sprintf(`
		ORDER BY creado_en DESC
		LIMIT $%d OFFSET $%d
	`, indice, indice+1)
	argumentos = append(argumentos, filtro.TamanoPagina, desplazamiento)

	filas, err := a.pool.Query(contexto, consulta, argumentos...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	usuarios := make([]UsuarioExterno, 0)
	for filas.Next() {
		u := UsuarioExterno{}
		if err := filas.Scan(&u.IdExterno, &u.CorreoElectronico, &u.PasswordHash, &u.Estado); err != nil {
			return nil, err
		}
		usuarios = append(usuarios, u)
	}
	return &ResultadoBusquedaUsuarios{
		Usuarios:       usuarios,
		TotalRegistros: total,
		Pagina:         filtro.Pagina,
		TamanoPagina:   filtro.TamanoPagina,
	}, nil
}

func (a *AdaptadorCodeplexVentas) ConsultarUsuario(contexto context.Context, idExterno string) (*UsuarioExterno, error) {
	u := &UsuarioExterno{}
	err := a.pool.QueryRow(contexto, `
		SELECT id::TEXT, correo_electronico, coalesce(password_hash,''), estado
		FROM usuario
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, idExterno).Scan(&u.IdExterno, &u.CorreoElectronico, &u.PasswordHash, &u.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUsuarioExternoNoEncontrado
	}
	return u, err
}

func (a *AdaptadorCodeplexVentas) ProvisionarUsuario(contexto context.Context, datos DatosProvisionar) (*UsuarioExterno, error) {
	if !validarEntradaProvisioning(datos) {
		return nil, errors.New("datos de provisioning incompletos")
	}

	hashDestino, err := cripto.HashearPasswordConArgon2id(datos.PasswordPlana)
	if err != nil {
		return nil, err
	}

	u := &UsuarioExterno{}
	err = a.pool.QueryRow(contexto, `
		INSERT INTO usuario (
			correo_electronico, password_hash, correo_electronico_verificado,
			estado, intentos_fallidos
		) VALUES ($1, $2, true, 'ACTIVO', 0)
		RETURNING id::TEXT, correo_electronico, estado
	`,
		strings.ToLower(strings.TrimSpace(datos.CorreoElectronico)),
		hashDestino,
	).Scan(&u.IdExterno, &u.CorreoElectronico, &u.Estado)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (a *AdaptadorCodeplexVentas) EditarUsuario(contexto context.Context, idExterno string, datos DatosEditarUsuarioExterno) (*UsuarioExterno, error) {
	if strings.TrimSpace(datos.CorreoElectronicoNuevo) == "" {
		return nil, errors.New("correo electrónico requerido")
	}
	u := &UsuarioExterno{}
	err := a.pool.QueryRow(contexto, `
		UPDATE usuario
		SET correo_electronico = $2, actualizado_en = now()
		WHERE id = $1 AND estado != 'ELIMINADO'
		RETURNING id::TEXT, correo_electronico, estado
	`,
		idExterno,
		strings.ToLower(strings.TrimSpace(datos.CorreoElectronicoNuevo)),
	).Scan(&u.IdExterno, &u.CorreoElectronico, &u.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUsuarioExternoNoEncontrado
	}
	return u, err
}

func (a *AdaptadorCodeplexVentas) CambiarPasswordUsuario(contexto context.Context, idExterno string, datos DatosCambiarPasswordExterno) error {
	if datos.PasswordNueva == "" {
		return errors.New("nueva contraseña requerida")
	}
	hashNuevo, err := cripto.HashearPasswordConArgon2id(datos.PasswordNueva)
	if err != nil {
		return err
	}
	tag, err := a.pool.Exec(contexto, `
		UPDATE usuario
		SET password_hash = $2, intentos_fallidos = 0, bloqueado_hasta = NULL, actualizado_en = now()
		WHERE id = $1 AND estado != 'ELIMINADO'
	`, idExterno, hashNuevo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUsuarioExternoNoEncontrado
	}
	return nil
}

func (a *AdaptadorCodeplexVentas) Diagnosticar(contexto context.Context) (*Diagnostico, error) {
	return diagnosticarPool(contexto, a.pool, "codeplex_ventas")
}

func validarEntradaProvisioning(datos DatosProvisionar) bool {
	return strings.TrimSpace(datos.CorreoElectronico) != "" &&
		strings.TrimSpace(datos.PasswordPlana) != ""
}
