package catalogo_sistemas

import "errors"

var (
	ErrCodigoRequerido             = errors.New("el código del sistema es requerido")
	ErrNombreRequerido             = errors.New("el nombre del sistema es requerido")
	ErrUrlAccesoRequerida          = errors.New("la URL de acceso es requerida")
	ErrMotorInvalido               = errors.New("motor de base de datos inválido")
	ErrAdaptadorRequerido          = errors.New("la clave del adaptador es requerida")
	ErrSistemaNoEncontrado         = errors.New("sistema no encontrado")
	ErrCodigoSistemaDuplicado      = errors.New("ya existe un sistema con ese código")
	ErrConexionLecturaNoEncontrada = errors.New("conexión de lectura no encontrada")
)
