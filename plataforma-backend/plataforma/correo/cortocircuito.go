package correo

import (
	"sync"
	"time"
)

// Cortocircuito: corta intentos cuando el servidor SMTP empieza a fallar.
// Razón: el proveedor (mail.codeplex.pe) bloquea la cuenta tras 5 fallos
// consecutivos. Pausamos a los 3 para mantenernos lejos del umbral.
//
// Ciclo:
//   - 0..2 errores seguidos → seguimos intentando.
//   - 3 errores seguidos    → pausamos N minutos, reseteamos contador a 0
//                             para no quemar la cuenta si la causa fue una
//                             ráfaga transitoria.
//   - Un éxito en cualquier momento → contador a 0.
type Cortocircuito struct {
	mu              sync.Mutex
	erroresSeguidos int
	pausadoHasta    time.Time
	limite          int
	duracionPausa   time.Duration
}

func NuevoCortocircuito(limite int, duracionPausa time.Duration) *Cortocircuito {
	return &Cortocircuito{limite: limite, duracionPausa: duracionPausa}
}

// PuedeIntentar retorna true si no estamos en período de pausa.
func (c *Cortocircuito) PuedeIntentar() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Now().After(c.pausadoHasta)
}

func (c *Cortocircuito) RegistrarExito() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.erroresSeguidos = 0
}

func (c *Cortocircuito) RegistrarError() (pausado bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.erroresSeguidos++
	if c.erroresSeguidos >= c.limite {
		c.pausadoHasta = time.Now().Add(c.duracionPausa)
		c.erroresSeguidos = 0
		return true
	}
	return false
}
