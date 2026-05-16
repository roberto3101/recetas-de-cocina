package errores

const (
	CodigoCredencialesInvalidas       = "CREDENCIALES_INVALIDAS"
	CodigoUsuarioBloqueado            = "USUARIO_BLOQUEADO"
	CodigoUsuarioInactivo             = "USUARIO_INACTIVO"
	CodigoUsuarioNoEncontrado         = "USUARIO_NO_ENCONTRADO"
	CodigoCorreoNoVerificado          = "CORREO_NO_VERIFICADO"
	CodigoCorreoYaRegistrado          = "CORREO_YA_REGISTRADO"

	CodigoSesionExpirada              = "SESION_EXPIRADA"
	CodigoSesionRevocada              = "SESION_REVOCADA"
	CodigoSesionRequerida             = "SESION_REQUERIDA"
	CodigoSesionNoEncontrada          = "SESION_NO_ENCONTRADA"
	CodigoSegundoFactorPendiente      = "SEGUNDO_FACTOR_PENDIENTE"
	CodigoSegundoFactorInvalido       = "SEGUNDO_FACTOR_INVALIDO"

	CodigoTotpYaActivo                = "TOTP_YA_ACTIVO"
	CodigoTotpNoActivo                = "TOTP_NO_ACTIVO"
	CodigoTotpCodigoInvalido          = "TOTP_CODIGO_INVALIDO"

	CodigoEntradaNoEncontrada         = "ENTRADA_NO_ENCONTRADA"
	CodigoEntradaInvalida             = "ENTRADA_INVALIDA"
	CodigoCifradoFallido              = "CIFRADO_FALLIDO"
	CodigoDescifradoFallido           = "DESCIFRADO_FALLIDO"

	CodigoSistemaNoEncontrado         = "SISTEMA_NO_ENCONTRADO"
	CodigoSistemaInvalido             = "SISTEMA_INVALIDO"
	CodigoCodigoSistemaDuplicado      = "CODIGO_SISTEMA_DUPLICADO"
	CodigoConexionExternaFallida      = "CONEXION_EXTERNA_FALLIDA"
	CodigoAdaptadorNoSoportado        = "ADAPTADOR_NO_SOPORTADO"
	CodigoAutoregistroNoPermitido     = "AUTOREGISTRO_NO_PERMITIDO"

	CodigoPeticionMalFormada          = "PETICION_MAL_FORMADA"
	CodigoValidacionFallida           = "VALIDACION_FALLIDA"
	CodigoErrorInterno                = "ERROR_INTERNO"
	CodigoLimiteVelocidadExcedido     = "LIMITE_VELOCIDAD_EXCEDIDO"
	CodigoPermisoDenegado             = "PERMISO_DENEGADO"
)
