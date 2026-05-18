import { useCallback, useEffect, useMemo, useState, FormEvent } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  AccesoGuardado,
  ListadoAccesos,
  ListadoSistemas,
  SistemaDisponible,
} from "@/capacidades/accesos/tipos";

type CredencialAutofill = {
  usuario: string;
  password: string;
  url_acceso: string;
  url_login: string;
};

export default function PaginaAcceso() {
  const navegar = useNavigate();
  const [texto, setTexto] = useState("");
  const [sistemaIdFiltro, setSistemaIdFiltro] = useState<string>("");

  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [accesos, setAccesos] = useState<AccesoGuardado[]>([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<{ titulo: string; usuario: string; password: string; copiada: boolean } | null>(null);

  const [accesoViendo, setAccesoViendo] = useState<AccesoGuardado | null>(null);
  const [accesoEditando, setAccesoEditando] = useState<AccesoGuardado | null>(null);

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

  async function abrirYCopiar(a: AccesoGuardado) {
    setError(null);
    setToast(null);
    const pestana = window.open("about:blank", "_blank");
    try {
      const cred = await pedirJson<CredencialAutofill>(`/cocina/boveda/accesos/${a.id}/bookmarklet`);
      let copiada = false;
      try {
        await navigator.clipboard.writeText(cred.password);
        copiada = true;
      } catch {
        copiada = false;
      }
      if (pestana) {
        pestana.location.href = cred.url_acceso || cred.url_login;
      }
      setToast({ titulo: a.titulo || "acceso", usuario: cred.usuario, password: cred.password, copiada });
    } catch (e) {
      if (pestana) pestana.close();
      if (e instanceof Error) setError(e.message);
    }
  }

  async function copiarTexto(t: string) {
    try {
      await navigator.clipboard.writeText(t);
    } catch {
      // ignore
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

      <form className="grid grid-cols-1 gap-3 sm:grid-cols-[2fr_1fr]" onSubmit={(e) => e.preventDefault()}>
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

      {toast && <ToastAutofill toast={toast} alCerrar={() => setToast(null)} alCopiar={copiarTexto} />}

      {accesoViendo && (
        <ModalVerDetalles
          acceso={accesoViendo}
          sistema={sistemasPorId.get(accesoViendo.sistema_destino_id)}
          alCerrar={() => setAccesoViendo(null)}
        />
      )}
      {accesoEditando && (
        <ModalEditarAcceso
          acceso={accesoEditando}
          sistemas={sistemas}
          alCerrar={() => setAccesoEditando(null)}
          alExito={() => { setAccesoEditando(null); void cargar(); }}
        />
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
                    <div className="flex flex-wrap justify-end items-center gap-x-3 gap-y-1">
                      <button onClick={() => setAccesoViendo(a)} className="text-xs text-cocina-marron hover:underline">Ver</button>
                      <button onClick={() => setAccesoEditando(a)} className="text-xs text-cocina-marron hover:underline">Editar</button>
                      <button onClick={() => eliminar(a)} className="text-xs text-red-700 hover:underline">Eliminar</button>
                    </div>
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

function ToastAutofill({ toast, alCerrar, alCopiar }: { toast: { titulo: string; usuario: string; password: string; copiada: boolean }; alCerrar: () => void; alCopiar: (t: string) => Promise<void> }) {
  const [pwdVisible, setPwdVisible] = useState(false);
  return (
    <div className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900 space-y-2">
      <div className="flex items-start gap-2">
        <span className="text-lg">✅</span>
        <div className="flex-1">
          <strong>{toast.titulo}</strong> abierto en pestaña nueva.
          {toast.copiada
            ? <span className="ml-2 text-xs text-emerald-700">Clave copiada — Ctrl+V en el form.</span>
            : <span className="ml-2 text-xs text-amber-700">⚠️ Clipboard bloqueado. Cópiala manual con el botón abajo.</span>}
        </div>
        <button onClick={alCerrar} className="text-stone-500 hover:text-stone-700" aria-label="Cerrar">✕</button>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
        <div className="flex items-center gap-2 bg-white rounded px-2 py-1 border border-emerald-100">
          <span className="text-stone-500 shrink-0">usuario:</span>
          <span className="font-mono break-all flex-1">{toast.usuario}</span>
          <button onClick={() => void alCopiar(toast.usuario)} className="text-emerald-700 shrink-0" title="copiar usuario">📋</button>
        </div>
        <div className="flex items-center gap-2 bg-white rounded px-2 py-1 border border-emerald-100">
          <span className="text-stone-500 shrink-0">clave:</span>
          <span className="font-mono break-all flex-1">
            {pwdVisible ? toast.password : "•".repeat(Math.min(12, toast.password.length))}
          </span>
          <button onClick={() => setPwdVisible(!pwdVisible)} className="text-emerald-700 shrink-0" title={pwdVisible ? "ocultar" : "ver"}>{pwdVisible ? "🙈" : "👁️"}</button>
          <button onClick={() => void alCopiar(toast.password)} className="text-emerald-700 shrink-0" title="copiar clave">📋</button>
        </div>
      </div>
    </div>
  );
}

function ModalContenedor({ titulo, alCerrar, children }: { titulo: string; alCerrar: () => void; children: React.ReactNode }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={alCerrar}>
      <div className="w-full max-w-lg rounded-md bg-white shadow-lg" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-stone-200 px-5 py-3">
          <h3 className="text-lg font-semibold text-cocina-oscuro">{titulo}</h3>
          <button type="button" onClick={alCerrar} className="text-stone-500 hover:text-cocina-marron" aria-label="Cerrar">✕</button>
        </div>
        <div className="px-5 py-4">{children}</div>
      </div>
    </div>
  );
}

function ModalVerDetalles({ acceso, sistema, alCerrar }: { acceso: AccesoGuardado; sistema?: SistemaDisponible; alCerrar: () => void }) {
  const [credencial, setCredencial] = useState<CredencialAutofill | null>(null);
  const [pwdVisible, setPwdVisible] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [cargando, setCargando] = useState(true);

  useEffect(() => {
    void (async () => {
      try {
        const r = await pedirJson<CredencialAutofill>(`/cocina/boveda/accesos/${acceso.id}/bookmarklet`);
        setCredencial(r);
      } catch (e) {
        if (e instanceof Error) setError(e.message);
      } finally {
        setCargando(false);
      }
    })();
  }, [acceso.id]);

  async function copiar(t: string) {
    try { await navigator.clipboard.writeText(t); } catch { /* ignore */ }
  }

  return (
    <ModalContenedor titulo="Detalles del acceso" alCerrar={alCerrar}>
      <dl className="space-y-2 text-sm">
        <Fila etiqueta="Título" valor={acceso.titulo || "—"} />
        <Fila etiqueta="Sistema" valor={sistema?.nombre || "—"} />
        <Fila etiqueta="URL" valor={sistema?.url_acceso || "—"} />
        <Fila etiqueta="Usuario" valor={acceso.usuario_externo} accion={() => void copiar(acceso.usuario_externo)} />
        <div className="grid grid-cols-[110px_1fr_auto_auto] gap-2 items-center">
          <dt className="font-medium text-stone-500">Clave</dt>
          <dd className="break-all text-cocina-oscuro font-mono text-xs">
            {cargando ? "Cargando…" : error ? <span className="text-red-700">{error}</span> :
              pwdVisible ? credencial?.password : "•".repeat(Math.min(20, credencial?.password.length ?? 8))}
          </dd>
          <button onClick={() => setPwdVisible(!pwdVisible)} className="text-stone-500 hover:text-cocina-marron text-sm" title="Ver/ocultar">{pwdVisible ? "🙈" : "👁️"}</button>
          <button onClick={() => credencial && void copiar(credencial.password)} className="text-stone-500 hover:text-cocina-marron text-sm" title="Copiar">📋</button>
        </div>
        {acceso.observaciones && <Fila etiqueta="Notas" valor={acceso.observaciones} />}
      </dl>
      <div className="mt-5 flex justify-end">
        <button type="button" onClick={alCerrar} className="boton-secundario">Cerrar</button>
      </div>
    </ModalContenedor>
  );
}

function Fila({ etiqueta, valor, accion }: { etiqueta: string; valor: string; accion?: () => void }) {
  return (
    <div className="grid grid-cols-[110px_1fr_auto] gap-2 items-center">
      <dt className="font-medium text-stone-500">{etiqueta}</dt>
      <dd className="break-all text-cocina-oscuro text-xs">{valor}</dd>
      {accion && <button onClick={accion} className="text-stone-500 hover:text-cocina-marron text-sm" title="Copiar">📋</button>}
    </div>
  );
}

function ModalEditarAcceso({ acceso, sistemas, alCerrar, alExito }: { acceso: AccesoGuardado; sistemas: SistemaDisponible[]; alCerrar: () => void; alExito: () => void }) {
  const [titulo, setTitulo] = useState(acceso.titulo);
  const [sistemaId, setSistemaId] = useState(acceso.sistema_destino_id);
  const [usuario, setUsuario] = useState(acceso.usuario_externo);
  const [claveNueva, setClaveNueva] = useState("");
  const [observaciones, setObservaciones] = useState(acceso.observaciones);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function alGuardar(e: FormEvent) {
    e.preventDefault();
    setError(null);
    if (!titulo.trim() || !usuario.trim() || !sistemaId) {
      setError("Título, usuario y sistema son obligatorios");
      return;
    }
    setEnviando(true);
    try {
      const cuerpo: Record<string, unknown> = {
        titulo: titulo.trim(),
        sistema_destino_id: sistemaId,
        usuario_externo: usuario.trim(),
        observaciones: observaciones.trim(),
      };
      if (claveNueva) cuerpo.password = claveNueva;
      await enviarJson(`/cocina/boveda/accesos/${acceso.id}`, cuerpo, "PUT");
      alExito();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <ModalContenedor titulo="Editar acceso" alCerrar={alCerrar}>
      <form onSubmit={alGuardar} className="space-y-3 text-sm">
        <div>
          <label htmlFor="edit-titulo" className="etiqueta-campo">Título</label>
          <input id="edit-titulo" type="text" className="campo-texto" value={titulo} onChange={(e) => setTitulo(e.target.value)} required maxLength={200} />
        </div>
        <div>
          <label htmlFor="edit-sistema" className="etiqueta-campo">Sistema</label>
          <select id="edit-sistema" className="campo-texto" value={sistemaId} onChange={(e) => setSistemaId(e.target.value)} required>
            <option value="">— Selecciona —</option>
            {sistemas.map((s) => (
              <option key={s.id} value={s.id}>{s.nombre} — {s.url_acceso}</option>
            ))}
          </select>
        </div>
        <div>
          <label htmlFor="edit-usuario" className="etiqueta-campo">Usuario</label>
          <input id="edit-usuario" type="text" className="campo-texto" value={usuario} onChange={(e) => setUsuario(e.target.value)} required maxLength={200} />
        </div>
        <div>
          <label htmlFor="edit-clave" className="etiqueta-campo">Nueva clave <span className="text-stone-400">(opcional, deja vacío para mantener la actual)</span></label>
          <input id="edit-clave" type="password" className="campo-texto" value={claveNueva} onChange={(e) => setClaveNueva(e.target.value)} autoComplete="new-password" maxLength={500} />
        </div>
        <div>
          <label htmlFor="edit-obs" className="etiqueta-campo">Observaciones</label>
          <textarea id="edit-obs" rows={2} className="campo-texto resize-y" value={observaciones} onChange={(e) => setObservaciones(e.target.value)} maxLength={1000} />
        </div>
        {error && <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-red-800">{error}</div>}
        <div className="flex justify-end gap-3 pt-2">
          <button type="button" className="boton-secundario" onClick={alCerrar} disabled={enviando}>Cancelar</button>
          <button type="submit" className="boton-primario" disabled={enviando}>{enviando ? "Guardando…" : "Guardar cambios"}</button>
        </div>
      </form>
    </ModalContenedor>
  );
}
