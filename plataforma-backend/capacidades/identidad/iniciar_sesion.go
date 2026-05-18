package identidad

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sistemas-unificados/compartido/validaciones"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/cripto"
	"sistemas-unificados/plataforma/gobierno/auditoria"
)

// Tolerante: 10 fallos antes de bloqueo, y solo 5 min de espera.
// Un humano normal no falla 10 veces; un bot brute force igual queda parado.
const MaxIntentosFallidosAntesDeBloqueo = 10
const MinutosBloqueoTrasFallos = 5
const HorasVidaSesion = 8

type DatosIniciarSesion struct {
	CorreoElectronico string
	PasswordPlana     string
	DispositivoId     string
	IpOrigen          string
	AgenteUsuario     string
}

type ResultadoIniciarSesion struct {
	TokenPlano            string
	RefreshTokenPlano     string
	SesionId              uuid.UUID
	UsuarioId             uuid.UUID
	ExpiraEn              time.Time
	RequiereSegundoFactor bool
}

func IniciarSesion(contexto context.Context, conexion *cockroach.ConexionBaseDatos, datos DatosIniciarSesion) (*ResultadoIniciarSesion, error) {
	correo := validaciones.NormalizarCorreo(datos.CorreoElectronico)
	if !validaciones.EsCorreoElectronicoValido(correo) {
		return nil, ErrCredencialesInvalidas
	}

	usuario, err := ConsultarUsuarioPorCorreo(contexto, conexion.Pool(), correo)
	if err != nil {
		if errors.Is(err, ErrUsuarioNoEncontrado) {
			return nil, ErrCredencialesInvalidas
		}
		return nil, err
	}

	if usuario.Estado == EstadoUsuarioEliminado || usuario.Estado == EstadoUsuarioInactivo {
		return nil, ErrUsuarioInactivo
	}
	if usuario.Estado == EstadoUsuarioBloqueado {
		if usuario.BloqueadoHasta == nil || time.Now().Before(*usuario.BloqueadoHasta) {
			return nil, ErrUsuarioBloqueado
		}
	}
	if !usuario.CorreoElectronicoVerificado && usuario.Estado == EstadoUsuarioPendiente {
		return nil, ErrCorreoNoVerificado
	}

	coincide, err := cripto.VerificarPasswordContraHashArgon2id(datos.PasswordPlana, usuario.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !coincide {
		_ = RegistrarIntentoFallido(contexto, conexion.Pool(), usuario.Id, MaxIntentosFallidosAntesDeBloqueo, MinutosBloqueoTrasFallos)
		return nil, ErrCredencialesInvalidas
	}

	tokenPlano, tokenHash, err := cripto.GenerarTokenSesionAleatorio()
	if err != nil {
		return nil, err
	}
	refreshPlano, refreshHash, err := cripto.GenerarTokenSesionAleatorio()
	if err != nil {
		return nil, err
	}

	_, errTotp := ConsultarTotpActivoDeUsuario(contexto, conexion.Pool(), usuario.Id)
	hayTotpActivo := errTotp == nil

	estadoSesion := EstadoSesionActiva
	if hayTotpActivo {
		estadoSesion = EstadoSesionPendienteSegundoFactor
	}

	dispositivoId := datos.DispositivoId
	if dispositivoId == "" {
		dispositivoId = uuid.NewString()
	}

	sesion := &SesionGlobal{
		UsuarioId:             usuario.Id,
		TokenHash:             tokenHash,
		RefreshTokenHash:      refreshHash,
		DispositivoId:         dispositivoId,
		IpOrigen:              datos.IpOrigen,
		AgenteUsuario:         datos.AgenteUsuario,
		SegundoFactorValidado: !hayTotpActivo,
		ExpiraEn:              time.Now().Add(HorasVidaSesion * time.Hour),
		Estado:                estadoSesion,
	}

	trans, err := conexion.Pool().Begin(contexto)
	if err != nil {
		return nil, err
	}
	defer trans.Rollback(contexto)

	if err := InsertarSesion(contexto, trans, sesion); err != nil {
		return nil, err
	}

	if !hayTotpActivo {
		if err := ReiniciarIntentosFallidosYMarcarInicio(contexto, trans, usuario.Id); err != nil {
			return nil, err
		}
	}

	if err := auditoria.RegistrarAccion(contexto, trans, auditoria.EntradaAuditoria{
		UsuarioId:     &usuario.Id,
		SesionId:      &sesion.Id,
		Modulo:        "SESIONES",
		Accion:        "SESION_INICIADA",
		Entidad:       "sesion_global",
		EntidadId:     &sesion.Id,
		Detalle:       estadoSesion,
		IpOrigen:      datos.IpOrigen,
		AgenteUsuario: datos.AgenteUsuario,
	}); err != nil {
		return nil, err
	}

	if err := trans.Commit(contexto); err != nil {
		return nil, err
	}

	return &ResultadoIniciarSesion{
		TokenPlano:            tokenPlano,
		RefreshTokenPlano:     refreshPlano,
		SesionId:              sesion.Id,
		UsuarioId:             usuario.Id,
		ExpiraEn:              sesion.ExpiraEn,
		RequiereSegundoFactor: hayTotpActivo,
	}, nil
}
