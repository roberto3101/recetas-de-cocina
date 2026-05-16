package adaptadores

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ClavePanelControl = "panel_control_facturacion_2026"

type AdaptadorPanelControl struct {
	pool *pgxpool.Pool
}

func NuevoAdaptadorPanelControl(pool *pgxpool.Pool) *AdaptadorPanelControl {
	return &AdaptadorPanelControl{pool: pool}
}

func (a *AdaptadorPanelControl) Clave() string {
	return ClavePanelControl
}

func (a *AdaptadorPanelControl) ProbarConexion(contexto context.Context) error {
	return a.pool.Ping(contexto)
}

func (a *AdaptadorPanelControl) ListarUsuarios(contexto context.Context, filtro FiltroBusquedaUsuarios) (*ResultadoBusquedaUsuarios, error) {
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
		condiciones = append(condiciones,
			fmt.Sprintf("(lower(coalesce(correo,'')) LIKE $%d OR lower(coalesce(nombres,'')) LIKE $%d)", indice, indice))
		argumentos = append(argumentos, "%"+strings.ToLower(strings.TrimSpace(filtro.Texto))+"%")
		indice++
	}
	if filtro.Estado == "ACTIVO" {
		condiciones = append(condiciones, "esta_activo = true")
	} else if filtro.Estado == "INACTIVO" {
		condiciones = append(condiciones, "esta_activo = false")
	}

	donde := ""
	if len(condiciones) > 0 {
		donde = "WHERE " + strings.Join(condiciones, " AND ")
	}

	var total int64
	if err := a.pool.QueryRow(contexto,
		"SELECT count(*) FROM usuarios_roles "+donde, argumentos...,
	).Scan(&total); err != nil {
		return nil, err
	}

	consulta := `
		SELECT id_usuario::TEXT, coalesce(correo,''), coalesce(nombres,''), CASE WHEN esta_activo THEN 'ACTIVO' ELSE 'INACTIVO' END
		FROM usuarios_roles ` + donde + fmt.Sprintf(`
		ORDER BY fecha_creacion DESC NULLS LAST
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
		if err := filas.Scan(&u.IdExterno, &u.CorreoElectronico, &u.NombreCompleto, &u.Estado); err != nil {
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

func (a *AdaptadorPanelControl) ConsultarUsuario(contexto context.Context, idExterno string) (*UsuarioExterno, error) {
	u := &UsuarioExterno{}
	err := a.pool.QueryRow(contexto, `
		SELECT id_usuario::TEXT, coalesce(correo,''), coalesce(nombres,''), CASE WHEN esta_activo THEN 'ACTIVO' ELSE 'INACTIVO' END
		FROM usuarios_roles
		WHERE id_usuario = $1
	`, idExterno).Scan(&u.IdExterno, &u.CorreoElectronico, &u.NombreCompleto, &u.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUsuarioExternoNoEncontrado
	}
	return u, err
}

func (a *AdaptadorPanelControl) ProvisionarUsuario(contexto context.Context, datos DatosProvisionar) (*UsuarioExterno, error) {
	return nil, errors.New("provisioning directo a panel_control no soportado: requiere usuario previo en login_global")
}

func (a *AdaptadorPanelControl) EditarUsuario(contexto context.Context, idExterno string, datos DatosEditarUsuarioExterno) (*UsuarioExterno, error) {
	u := &UsuarioExterno{}
	err := a.pool.QueryRow(contexto, `
		UPDATE usuarios_roles
		SET correo = COALESCE(NULLIF($2, ''), correo),
		    nombres = COALESCE(NULLIF($3, ''), nombres),
		    fecha_actualizacion = now()
		WHERE id_usuario = $1
		RETURNING id_usuario::TEXT, coalesce(correo,''), coalesce(nombres,''), CASE WHEN esta_activo THEN 'ACTIVO' ELSE 'INACTIVO' END
	`, idExterno,
		strings.ToLower(strings.TrimSpace(datos.CorreoElectronicoNuevo)),
		strings.TrimSpace(datos.NombreCompletoNuevo),
	).Scan(&u.IdExterno, &u.CorreoElectronico, &u.NombreCompleto, &u.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUsuarioExternoNoEncontrado
	}
	return u, err
}

func (a *AdaptadorPanelControl) CambiarPasswordUsuario(contexto context.Context, idExterno string, datos DatosCambiarPasswordExterno) error {
	return errors.New("panel_control no almacena contraseña: cambia la del usuario en login_global")
}

func (a *AdaptadorPanelControl) Diagnosticar(contexto context.Context) (*Diagnostico, error) {
	return diagnosticarPool(contexto, a.pool, "panel_control_facturacion_2026")
}
