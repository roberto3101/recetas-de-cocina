package identidad

import "errors"

var (
	// Usuarios
	ErrCorreoInvalido         = errors.New("correo electrónico inválido")
	ErrCorreoYaRegistrado     = errors.New("ya existe un usuario con ese correo")
	ErrUsuarioNoEncontrado    = errors.New("usuario no encontrado")
	ErrUsuarioInactivo        = errors.New("usuario inactivo")
	ErrUsuarioBloqueado       = errors.New("usuario bloqueado")
	ErrUsuarioEliminado       = errors.New("usuario eliminado")
	ErrPasswordActualInvalido = errors.New("la contraseña actual no es correcta")

	// Sesiones
	ErrCredencialesInvalidas  = errors.New("credenciales inválidas")
	ErrCorreoNoVerificado     = errors.New("correo no verificado")
	ErrSesionNoEncontrada     = errors.New("sesión no encontrada")
	ErrSesionRevocada         = errors.New("sesión revocada")
	ErrSesionExpirada         = errors.New("sesión expirada")
	ErrSegundoFactorPendiente = errors.New("segundo factor pendiente de validación")

	// Verificacion correo
	ErrVerificacionNoEncontrada   = errors.New("verificación de correo no encontrada o ya consumida")
	ErrCodigoRecuperacionInvalido = errors.New("código de recuperación inválido o expirado")

	// TOTP
	ErrTotpYaActivo       = errors.New("el usuario ya tiene TOTP activo")
	ErrTotpNoActivo       = errors.New("el usuario no tiene TOTP activo")
	ErrTotpNoEncontrado   = errors.New("activación TOTP no encontrada o ya consumida")
	ErrCodigoTotpInvalido = errors.New("código TOTP inválido")
)
