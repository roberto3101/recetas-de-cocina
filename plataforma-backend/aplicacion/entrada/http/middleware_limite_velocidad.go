package http

import (
	"net/http"
	"sync"
	"time"

	"sistemas-unificados/plataforma/gobierno/errores"
)

type limitadorVelocidadEnMemoria struct {
	mu                  sync.Mutex
	hitsPorClienteIp    map[string][]time.Time
	maxIntentosVentana  int
	ventana             time.Duration
}

func NuevoLimitadorVelocidad(maxIntentos int, ventana time.Duration) *limitadorVelocidadEnMemoria {
	return &limitadorVelocidadEnMemoria{
		hitsPorClienteIp:   make(map[string][]time.Time),
		maxIntentosVentana: maxIntentos,
		ventana:            ventana,
	}
}

func (l *limitadorVelocidadEnMemoria) Permitir(clienteIp string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	ahora := time.Now()
	corte := ahora.Add(-l.ventana)

	hits := l.hitsPorClienteIp[clienteIp]
	vigentes := hits[:0]
	for _, h := range hits {
		if h.After(corte) {
			vigentes = append(vigentes, h)
		}
	}

	if len(vigentes) >= l.maxIntentosVentana {
		l.hitsPorClienteIp[clienteIp] = vigentes
		return false
	}

	vigentes = append(vigentes, ahora)
	l.hitsPorClienteIp[clienteIp] = vigentes
	return true
}

func (l *limitadorVelocidadEnMemoria) AplicarComoMiddleware() func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
			ip := obtenerIpRemota(peticion)
			if !l.Permitir(ip) {
				ResponderError(escritor, http.StatusTooManyRequests, errores.CodigoLimiteVelocidadExcedido, "demasiados intentos, intenta más tarde")
				return
			}
			siguiente.ServeHTTP(escritor, peticion)
		})
	}
}
