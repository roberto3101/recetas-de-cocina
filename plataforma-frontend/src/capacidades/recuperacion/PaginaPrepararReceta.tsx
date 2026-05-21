import { FormEvent, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import CampoContrasena from "@/plataforma/ui/CampoContrasena";
import MedidorPassword, { evaluarPassword } from "@/plataforma/ui/MedidorPassword";

// PaginaPrepararReceta: usuario llega aquí desde el link que se le envió.
// La URL es /receta/:codigo donde :codigo es el token aleatorio.
//
// Flujo:
//   1. Al montar, valida el token contra /buscar/validar-receta/:codigo
//   2. Si es válido → muestra form de nueva clave + confirmación
//   3. Al submit → POST /buscar/preparar-receta con codigo + nueva_clave
//   4. Éxito → redirige al login después de 2s
type EstadoValidacion =
  | { tipo: "cargando" }
  | { tipo: "ok"; correoEnmascarado: string }
  | { tipo: "invalido"; mensaje: string };

export default function PaginaPrepararReceta() {
  const { codigo } = useParams<{ codigo: string }>();
  const [estado, setEstado] = useState<EstadoValidacion>({ tipo: "cargando" });
  const [nuevaClave, setNuevaClave] = useState("");
  const [confirmacion, setConfirmacion] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [mensaje, setMensaje] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!codigo) {
      setEstado({ tipo: "invalido", mensaje: "Enlace incompleto." });
      return;
    }
    void (async () => {
      try {
        const r = await pedirJson<{ correo_enmascarado: string }>(
          `/buscar/validar-receta/${encodeURIComponent(codigo)}`,
        );
        setEstado({ tipo: "ok", correoEnmascarado: r.correo_enmascarado });
      } catch (e) {
        const mensaje =
          e instanceof ErrorApi || e instanceof Error
            ? e.message
            : "Enlace inválido.";
        setEstado({ tipo: "invalido", mensaje });
      }
    })();
  }, [codigo]);

  async function alEnviar(evento: FormEvent) {
    evento.preventDefault();
    setError(null);
    setMensaje(null);

    const evaluacion = evaluarPassword(nuevaClave);
    if (!evaluacion.todoOk) {
      setError("La nueva clave no cumple los requisitos.");
      return;
    }
    if (nuevaClave !== confirmacion) {
      setError("La confirmación no coincide.");
      return;
    }

    setEnviando(true);
    try {
      await enviarJson<{ password_restablecida: boolean }>(
        "/buscar/preparar-receta",
        { codigo, nueva_clave: nuevaClave },
      );
      setMensaje("Receta restablecida. Te llevaremos al inicio en un momento.");
      // Full page reload con window.location en vez de navegar(): tras un
      // cambio de password queremos arrancar limpio (SW nuevo, Context
      // fresco, sesión anónima sin restos). El navegar() de react-router
      // mantenía JS state stale que confundía al ProveedorSesion.
      setTimeout(() => { window.location.href = "/"; }, 2000);
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  if (estado.tipo === "cargando") {
    return (
      <main className="mx-auto max-w-md px-4 py-10">
        <p className="text-stone-500 text-sm">Verificando enlace…</p>
      </main>
    );
  }

  if (estado.tipo === "invalido") {
    return (
      <main className="mx-auto max-w-md px-4 py-10">
        <h1 className="text-3xl font-bold text-cocina-marron">Enlace inválido</h1>
        <p className="mt-2 text-cocina-oscuro text-sm">
          {estado.mensaje}. Si necesitas un nuevo enlace, pídelo nuevamente.
        </p>
        <div className="mt-5 flex gap-3">
          <Link to="/solicitar-receta" className="boton-primario">
            Pedir nuevo enlace
          </Link>
          <Link to="/" className="boton-secundario">
            Volver al inicio
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-md px-4 py-10">
      <h1 className="text-3xl font-bold text-cocina-marron">Reescribir receta</h1>
      <p className="mt-2 text-cocina-oscuro text-sm">
        Vas a cambiar la clave de la cuenta{" "}
        <span className="font-mono text-cocina-marron">{estado.correoEnmascarado}</span>.
      </p>

      {mensaje && (
        <div className="mt-5 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          {mensaje}
        </div>
      )}
      {error && (
        <div className="mt-5 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {error}
        </div>
      )}

      <form
        onSubmit={alEnviar}
        className="mt-6 rounded-md bg-cocina-fondo p-6 shadow-sm border border-amber-100 space-y-4"
      >
        <div>
          <label className="etiqueta-campo">Nueva receta secreta</label>
          <CampoContrasena
            required
            value={nuevaClave}
            onChange={(e) => setNuevaClave(e.target.value)}
            autoComplete="new-password"
            validez={
              nuevaClave.length === 0
                ? undefined
                : evaluarPassword(nuevaClave).todoOk
                  ? "ok"
                  : "error"
            }
          />
          <MedidorPassword valor={nuevaClave} confirmacion={confirmacion} />
        </div>
        <div>
          <label className="etiqueta-campo">Confirmar receta</label>
          <CampoContrasena
            required
            value={confirmacion}
            onChange={(e) => setConfirmacion(e.target.value)}
            autoComplete="new-password"
            validez={
              confirmacion.length === 0
                ? undefined
                : confirmacion === nuevaClave && nuevaClave.length > 0
                  ? "ok"
                  : "error"
            }
          />
        </div>
        <button
          type="submit"
          className="boton-primario w-full"
          disabled={
            enviando ||
            !evaluarPassword(nuevaClave).todoOk ||
            nuevaClave !== confirmacion
          }
        >
          {enviando ? "Guardando…" : "Guardar nueva receta"}
        </button>
      </form>
    </main>
  );
}
