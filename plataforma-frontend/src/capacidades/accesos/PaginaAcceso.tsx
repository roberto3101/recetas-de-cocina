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
                    <a
                      href={`/cocina/boveda/accesos/${a.id}/autofill`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-medium text-cocina-marron hover:underline"
                      title="Abrir con login automático"
                    >
                      {a.titulo || s?.nombre || "(sin título)"} ↗
                    </a>
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
