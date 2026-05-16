package adaptadores

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProveedorPoolExterno func(contexto context.Context, idSistema uuid.UUID) (*pgxpool.Pool, error)

type RegistroAdaptadores struct {
	mu             sync.Mutex
	cachePorSistema map[uuid.UUID]Adaptador
	cachePools     map[uuid.UUID]*pgxpool.Pool
	proveedorPool  ProveedorPoolExterno
}

func NuevoRegistroAdaptadores(proveedor ProveedorPoolExterno) *RegistroAdaptadores {
	return &RegistroAdaptadores{
		cachePorSistema: make(map[uuid.UUID]Adaptador),
		cachePools:      make(map[uuid.UUID]*pgxpool.Pool),
		proveedorPool:   proveedor,
	}
}

func (r *RegistroAdaptadores) Resolver(claveAdaptador string, idSistemaDestino uuid.UUID) (Adaptador, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if adaptador, ok := r.cachePorSistema[idSistemaDestino]; ok {
		return adaptador, nil
	}

	contexto := context.Background()
	pool, err := r.proveedorPool(contexto, idSistemaDestino)
	if err != nil {
		return nil, err
	}
	r.cachePools[idSistemaDestino] = pool

	adaptador, err := construirAdaptador(claveAdaptador, pool)
	if err != nil {
		pool.Close()
		delete(r.cachePools, idSistemaDestino)
		return nil, err
	}

	r.cachePorSistema[idSistemaDestino] = adaptador
	return adaptador, nil
}

func (r *RegistroAdaptadores) CerrarTodo() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, pool := range r.cachePools {
		pool.Close()
		delete(r.cachePools, id)
		delete(r.cachePorSistema, id)
	}
}

func construirAdaptador(clave string, pool *pgxpool.Pool) (Adaptador, error) {
	switch clave {
	case ClaveCodeplexVentas:
		return NuevoAdaptadorCodeplexVentas(pool), nil
	case ClavePanelControl:
		return NuevoAdaptadorPanelControl(pool), nil
	}
	return nil, fmt.Errorf("adaptador desconocido: %s", clave)
}

var ErrSinProveedor = errors.New("no hay proveedor de conexiones externas configurado")
