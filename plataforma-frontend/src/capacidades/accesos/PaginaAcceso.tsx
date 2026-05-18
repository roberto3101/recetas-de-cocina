import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  AccesoGuardado,
  ListadoAccesos,
  ListadoSistemas,
  SistemaDisponible,
} from "@/capacidades/accesos/tipos";

export default function PaginaAcceso() {
  const navegar = useNavigate();
  const [texto, setTexto] = useState("");
  const [sistemaIdFiltro, setSistemaIdFiltro] = useState<string>("");

  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [accesos, setAccesos] = useState<AccesoGuardado[]>([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<{ titulo: string; usuario: string } | null>(null);

  const cargar = useCallback(async () => {
    setCargando(true);
    setError(null);
    try {
      const [sis, acc] = await Promise.all([
        pedirJson<ListadoSistemas>("/cocina/sistemas"),
        pedirJson<ListadoAccesos>("/cocina/boveda/accesos"),
      ]);
      setSistemas(sis.sistemas);
      setAccesos(acc.accesos);
    } catch (e) {
      if (e instanceof Error) setError(e.message);
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    void cargar();
  }, [cargar]);

  const sistemasPorId = useMemo(() => {
    const m = new Map<string, SistemaDisponible>();
    for (const s of sistemas) m.set(s.id, s);
    return m;
  }, [sistemas]);

  const accesosFiltrados = useMemo(() => {
    const t = texto.trim().toLowerCase();
    return accesos.filter((a) => {
      if (sistemaIdFiltro && a.sistema_destino_id !== sistemaIdFiltro) return false;
      if (!t) return true;
      const s = sistemasPorId.get(a.sistema_destino_id);
      return (
        a.titulo.toLowerCase().includes(t) ||
        a.usuario_externo.toLowerCase().includes(t) ||
        (s?.nombre.toLowerCase().includes(t) ?? false) ||
        (s?.url_acceso.toLowerCase().includes(t) ?? false)
      );
    });
  }, [accesos, sistemaIdFiltro, texto, sistemasPorId]);

  type CredencialAutofill = {
    usuario: string;
    password: string;
    url_acceso: string;
    url_login: string;
  };

  async function abrirYCopiar(a: AccesoGuardado) {
    setError(null);
    setToast(null);
    // Abrir pestaña ANTES del fetch para evitar bloqueo de popups en Safari/Chrome
    const pestana = window.open("about:blank", "_blank");
    try {
      const cred = await pedirJson<CredencialAutofill>(`/cocina/boveda/accesos/${a.id}/bookmarklet`);
      try {
        await navigator.clipboard.writeText(cred.password);
      } catch {
        // Fallback: no se pudo escribir clipboard (permiso negado). Mostrar pwd para copia manual.
      }
      if (pestana) {
        pestana.location.href = cred.url_acceso || cred.url_login;
      }
      setToast({ titulo: a.titulo || "acceso", usuario: cred.usuario });
      setTimeout(() => setToast(null), 12000);
    } catch (e) {
      if (pestana) pestana.close();
      if (e instanceof Error) setError(e.message);
    }
  }

  async function eliminar(a: AccesoGuardado) {
    if (!confirm(`¿Eliminar el acceso "${a.titulo || a.usuario_externo}"?`)) return;
    try {
      await enviarJson(`/cocina/boveda/accesos/${a.id}`, undefined, "DELETE");
      void cargar();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4 flex-wrap">
        <div>
          <h2 className="text-2xl font-semibold text-cocina-oscuro">Acceso</h2>
          <p className="text-sm text-stone-500">
            {cargando ? "Cargando…" : `${accesosFiltrados.length} de ${accesos.length} accesos`}
          </p>
        </div>
        <button onClick={() => navegar("/panel/registro")} className="boton-primario">+ Registrar acceso</button>
      </div>

      <form
        className="grid grid-cols-1 gap-3 sm:grid-cols-[2fr_1fr]"
        onSubmit={(e) => e.preventDefault()}
      >
        <input
          type="text"
          placeholder="Buscar por título, usuario o sistema…"
          className="campo-texto"
          value={texto}
          onChange={(e) => setTexto(e.target.value)}
        />
        <select
          className="campo-texto"
          value={sistemaIdFiltro}
          onChange={(e) => setSistemaIdFiltro(e.target.value)}
        >
          <option value="">Todas las URLs</option>
          {sistemas.map((s) => (
            <option key={s.id} value={s.id}>{s.nombre}</option>
          ))}
        </select>
      </form>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">{error}</div>
      )}

      {toast && (
        <div className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900 flex items-start gap-3 flex-wrap">
          <span className="text-lg">✅</span>
          <div className="flex-1">
            <div><strong>{toast.titulo}</strong> abierto en pestaña nueva.</div>
            <div className="text-xs mt-1">
              Usuario: <span className="font-mono">{toast.usuario}</span>
              <button
                type="button"
                onClick={() => navigator.clipboard.writeText(toast.usuario)}
                className="ml-2 text-emerald-700 underline"
              >copiar usuario</button>
            </div>
            <div className="text-xs">Clave: 📋 ya copiada — Ctrl+V (o ⌘+V) en el campo de contraseña.</div>
          </div>
          <button onClick={() => setToast(null)} className="text-stone-500 hover:text-stone-700" aria-label="Cerrar">✕</button>
        </div>
      )}

      <div className="overflow-x-auto rounded-md border border-stone-200">
        <table className="min-w-full text-sm">
          <thead className="bg-stone-100 text-left text-xs uppercase tracking-wider text-stone-600">
            <tr>
              <th className="w-12 px-3 py-2">#</th>
              <th className="px-3 py-2">Título / URL</th>
              <th className="px-3 py-2">Usuario</th>
              <th className="px-3 py-2 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody>
            {cargando && (
              <tr><td colSpan={4} className="px-3 py-6 text-center text-stone-500">Cargando…</td></tr>
            )}
            {!cargando && accesosFiltrados.length === 0 && (
              <tr><td colSpan={4} className="px-3 py-6 text-center text-stone-500">
                No hay accesos. Click en <strong>+ Registrar acceso</strong>.
              </td></tr>
            )}
            {!cargando && accesosFiltrados.map((a, idx) => {
              const s = sistemasPorId.get(a.sistema_destino_id);
              return (
                <tr key={a.id} className="border-t border-stone-100 hover:bg-cocina-fondo align-top">
                  <td className="px-3 py-2 text-stone-500">{idx + 1}</td>
                  <td className="px-3 py-2">
                    <button
                      type="button"
                      onClick={() => void abrirYCopiar(a)}
                      className="font-medium text-cocina-marron hover:underline text-left"
                      title="Abre el sistema y copia la clave al portapapeles"
                    >
                      {a.titulo || s?.nombre || "(sin título)"} ↗
                    </button>
                    <div className="text-xs text-stone-400 break-all">{s?.url_acceso ?? "—"}</div>
                  </td>
                  <td className="px-3 py-2 text-cocina-oscuro break-all">
                    {a.usuario_externo}
                    {a.observaciones && (
                      <div className="text-xs text-stone-400">{a.observaciones}</div>
                    )}
                  </td>
                  <td className="px-3 py-2 text-right whitespace-nowrap">
                    <button
                      type="button"
                      onClick={() => eliminar(a)}
                      className="text-xs text-red-700 hover:underline"
                    >Eliminar</button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
