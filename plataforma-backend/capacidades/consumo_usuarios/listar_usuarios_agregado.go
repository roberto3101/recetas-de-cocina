package consumo_usuarios

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"sistemas-unificados/capacidades/catalogo_sistemas"
	"sistemas-unificados/capacidades/consumo_usuarios/adaptadores"
	"sistemas-unificados/persistencia/cockroach"
)

type UsuarioConSistema struct {
	IdExterno         string            `json:"id_externo"`
	CorreoElectronico string            `json:"correo_electronico"`
	NombreCompleto    string            `json:"nombre_completo"`
	Estado            string            `json:"estado"`
	PasswordHash      string            `json:"password_hash,omitempty"`
	SistemaDestinoId  uuid.UUID         `json:"sistema_destino_id"`
	SistemaCodigo     string            `json:"sistema_codigo"`
	SistemaNombre     string            `json:"sistema_nombre"`
	SistemaUrlAcceso  string            `json:"sistema_url_acceso"`
	DatosAdicionales  map[string]string `json:"datos_adicionales,omitempty"`
}

type ErrorPorSistema struct {
	SistemaDestinoId uuid.UUID `json:"sistema_destino_id"`
	SistemaCodigo    string    `json:"sistema_codigo"`
	SistemaNombre    string    `json:"sistema_nombre"`
	Mensaje          string    `json:"mensaje"`
}

type ResultadoListadoAgregado struct {
	Usuarios          []UsuarioConSistema `json:"usuarios"`
	TotalRegistros   int64               `json:"total_registros"`
	ErroresPorSistema []ErrorPorSistema   `json:"errores_por_sistema,omitempty"`
}

type DatosListarAgregado struct {
	Filtro adaptadores.FiltroBusquedaUsuarios
}

func ListarUsuariosDeTodosLosSistemas(contexto context.Context, conexion *cockroach.ConexionBaseDatos, resolver adaptadores.Resolver, datos DatosListarAgregado) (*ResultadoListadoAgregado, error) {
	sistemas, err := catalogo_sistemas.ListarSistemasActivosDb(contexto, conexion.Pool())
	if err != nil {
		return nil, err
	}

	type respuesta struct {
		sistema  catalogo_sistemas.SistemaDestino
		usuarios []adaptadores.UsuarioExterno
		total    int64
		err      error
	}

	canal := make(chan respuesta, len(sistemas))
	var grupo sync.WaitGroup

	for _, sistema := range sistemas {
		if !sistema.SoportaLectura {
			continue
		}
		grupo.Add(1)
		go func(s catalogo_sistemas.SistemaDestino) {
			defer grupo.Done()
			adaptador, err := resolver.Resolver(s.ClaveAdaptador, s.Id)
			if err != nil {
				canal <- respuesta{sistema: s, err: err}
				return
			}
			resultado, err := adaptador.ListarUsuarios(contexto, datos.Filtro)
			if err != nil {
				canal <- respuesta{sistema: s, err: err}
				return
			}
			canal <- respuesta{sistema: s, usuarios: resultado.Usuarios, total: resultado.TotalRegistros}
		}(sistema)
	}

	grupo.Wait()
	close(canal)

	salida := &ResultadoListadoAgregado{Usuarios: []UsuarioConSistema{}}
	for r := range canal {
		if r.err != nil {
			salida.ErroresPorSistema = append(salida.ErroresPorSistema, ErrorPorSistema{
				SistemaDestinoId: r.sistema.Id,
				SistemaCodigo:    r.sistema.Codigo,
				SistemaNombre:    r.sistema.Nombre,
				Mensaje:          r.err.Error(),
			})
			continue
		}
		salida.TotalRegistros += r.total
		for _, u := range r.usuarios {
			salida.Usuarios = append(salida.Usuarios, UsuarioConSistema{
				IdExterno:         u.IdExterno,
				CorreoElectronico: u.CorreoElectronico,
				NombreCompleto:    u.NombreCompleto,
				Estado:            u.Estado,
				PasswordHash:      u.PasswordHash,
				SistemaDestinoId:  r.sistema.Id,
				SistemaCodigo:     r.sistema.Codigo,
				SistemaNombre:     r.sistema.Nombre,
				SistemaUrlAcceso:  r.sistema.UrlAcceso,
				DatosAdicionales:  u.DatosAdicionales,
			})
		}
	}

	return salida, nil
}
