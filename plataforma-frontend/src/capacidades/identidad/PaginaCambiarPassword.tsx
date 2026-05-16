import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson } from "@/plataforma/red/cliente_api";

export default function PaginaCambiarPassword() {
  const navegar = useNavigate();
  const [passwordActual, setPasswordActual] = useState("");
  const [passwordNueva, setPasswordNueva] = useState("");
  const [confirmacion, setConfirmacion] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [mensaje, setMensaje] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function alGuardar(evento: FormEvent) {
    evento.preventDefault();
    setError(null);
    setMensaje(null);

    if (passwordNueva !== confirmacion) {
      setError("La confirmación no coincide con la nueva contraseña");
      return;
    }

    setEnviando(true);
    try {
      await enviarJson<{ password_cambiado: boolean }>(
        "/cocina/identidad/cambiar-password",
        { password_actual: passwordActual, password_nueva: passwordNueva },
        "PUT"
      );
      setMensaje("Contraseña actualizada. Las sesiones anteriores fueron invalidadas.");
      setTimeout(() => navegar("/"), 1500);
    } catch (e) {
      if (e instanceof ErrorApi) setError(e.message);
      else if (e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div className="max-w-xl">
      <h2 className="text-2xl font-semibold text-cocina-oscuro mb-1">Cambiar contraseña</h2>
      <p className="text-sm text-stone-500 mb-5">
        Cambiar la contraseña cerrará todas tus sesiones activas. Tendrás que iniciar sesión de nuevo.
      </p>

      {mensaje && (
        <div className="mb-4 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-2 text-sm text-emerald-800">
          {mensaje}
        </div>
      )}
      {error && (
        <div className="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">
          {error}
        </div>
      )}

      <form onSubmit={alGuardar} className="space-y-4">
        <div>
          <label className="etiqueta-campo">Contraseña actual</label>
          <input
            type="password"
            required
            className="campo-texto"
            value={passwordActual}
            onChange={(e) => setPasswordActual(e.target.value)}
            autoComplete="current-password"
          />
        </div>
        <div>
          <label className="etiqueta-campo">Nueva contraseña</label>
          <input
            type="password"
            required
            className="campo-texto"
            value={passwordNueva}
            onChange={(e) => setPasswordNueva(e.target.value)}
            autoComplete="new-password"
          />
          <p className="mt-1 text-xs text-stone-500">
            Mínimo 12 caracteres, con mayúsculas, minúsculas, dígitos y al menos un especial.
          </p>
        </div>
        <div>
          <label className="etiqueta-campo">Confirmar nueva contraseña</label>
          <input
            type="password"
            required
            className="campo-texto"
            value={confirmacion}
            onChange={(e) => setConfirmacion(e.target.value)}
            autoComplete="new-password"
          />
        </div>
        <button type="submit" className="boton-primario" disabled={enviando}>
          {enviando ? "Guardando…" : "Cambiar contraseña"}
        </button>
      </form>
    </div>
  );
}
