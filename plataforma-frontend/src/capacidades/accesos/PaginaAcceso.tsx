import { useCallback, useEffect, useMemo, useState, FormEvent } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  AccesoGuardado,
  ETIQUETAS_TIPO,
  ListadoAccesos,
  ListadoSistemas,
  SistemaDisponible,
  TIPOS_ACCESO,
  TipoAcceso,
} from "@/capacidades/accesos/tipos";

type CredencialAutofill = {
  usuario: string;
  password: string;
  url_acceso: string;
  url_login: string;
};

type FiltroEstado = "ACTIVO" | "REVOCADO" | "TODOS";

export default function PaginaAcceso() {
  const navegar = useNavigate();
  const [texto, setTexto] = useState("");
  const [sistemaIdFiltro, setSistemaIdFiltro] = useState<string>("");
  const [filtroEstado, setFiltroEstado] = useState<FiltroEstado>("ACTIVO");

  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [accesos, setAccesos] = useState<AccesoGuardado[]>([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [pwdsVisibles, setPwdsVisibles] = useState<Record<string, boolean>>({});
  const [pwdsCache, setPwdsCache] = useState<Record<string, string>>({});
  const [mensajeFlash, setMensajeFlash] = useState<string | null>(null);

  const [accesoEditando, setAccesoEditando] = useState<AccesoGuardado | null>(null);
  const [accesoViendo, setAccesoViendo] = useState<AccesoGuardado | null>(null);

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

  useEffect(() => { void cargar(); }, [cargar]);

  const sistemasPorId = useMemo(() => {
    const m = new Map<string, SistemaDisponible>();
    for (const s of sistemas) m.set(s.id, s);
    return m;
  }, [sistemas]);

  const accesosFiltrados = useMemo(() => {
    const t = texto.trim().toLowerCase();
    return accesos.filter((a) => {
      if (filtroEstado !== "TODOS" && a.estado !== filtroEstado) return false;
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
  }, [accesos, sistemaIdFiltro, texto, filtroEstado, sistemasPorId]);

  function flash(msg: string) {
    setMensajeFlash(msg);
    setTimeout(() => setMensajeFlash((m) => (m === msg ? null : m)), 1800);
  }

  async function copiar(valor: string, etiqueta: string) {
    try {
      await navigator.clipboard.writeText(valor);
      flash(`${etiqueta} copiado`);
    } catch {
      flash("No se pudo copiar (clipboard bloqueado)");
    }
  }

  async function obtenerPwd(a: AccesoGuardado): Promise<string | null> {
    if (pwdsCache[a.id]) return pwdsCache[a.id];
    try {
      const cred = await pedirJson<CredencialAutofill>(`/cocina/boveda/accesos/${a.id}/bookmarklet`);
      setPwdsCache((c) => ({ ...c, [a.id]: cred.password }));
      return cred.password;
    } catch (e) {
      if (e instanceof Error) setError(e.message);
      return null;
    }
  }

  async function togglePwdVisible(a: AccesoGuardado) {
    if (!pwdsCache[a.id]) {
      const p = await obtenerPwd(a);
      if (!p) return;
    }
    setPwdsVisibles((v) => ({ ...v, [a.id]: !v[a.id] }));
  }

  async function copiarPwd(a: AccesoGuardado) {
    const p = pwdsCache[a.id] ?? (await obtenerPwd(a));
    if (p) void copiar(p, "Clave");
  }

  async function abrirEnPestana(a: AccesoGuardado) {
    setError(null);
    const pestana = window.open("about:blank", "_blank");
    try {
      const cred = await pedirJson<CredencialAutofill>(`/cocina/boveda/accesos/${a.id}/bookmarklet`);
      setPwdsCache((c) => ({ ...c, [a.id]: cred.password }));
      try { await navigator.clipboard.writeText(cred.password); flash("Clave copiada"); } catch { /* silent */ }
      if (pestana) pestana.location.href = cred.url_acceso || cred.url_login;
    } catch (e) {
      if (pestana) pestana.close();
      if (e instanceof Error) setError(e.message);
    }
  }

  async function desactivar(a: AccesoGuardado) {
    if (!confirm(`¿Desactivar "${a.titulo || a.usuario_externo}"? Lo podrás reactivar después.`)) return;
    try {
      await enviarJson(`/cocina/boveda/accesos/${a.id}`, undefined, "DELETE");
      void cargar();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    }
  }

  async function reactivar(a: AccesoGuardado) {
    try {
      await enviarJson(`/cocina/boveda/accesos/${a.id}/reactivar`, undefined, "POST");
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

      <form className="grid grid-cols-1 gap-3 sm:grid-cols-[2fr_1fr_1fr]" onSubmit={(e) => e.preventDefault()}>
        <input
          type="text"
          placeholder="Buscar por título, usuario o sistema…"
          className="campo-texto"
          value={texto}
          onChange={(e) => setTexto(e.target.value)}
        />
        <select className="campo-texto" value={sistemaIdFiltro} onChange={(e) => setSistemaIdFiltro(e.target.value)}>
          <option value="">Todas las URLs</option>
          {sistemas.map((s) => (
            <option key={s.id} value={s.id}>{s.nombre}</option>
          ))}
        </select>
        <select className="campo-texto" value={filtroEstado} onChange={(e) => setFiltroEstado(e.target.value as FiltroEstado)}>
          <option value="ACTIVO">Solo activos</option>
          <option value="REVOCADO">Solo inactivos</option>
          <option value="TODOS">Todos</option>
        </select>
      </form>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">{error}</div>
      )}

      {mensajeFlash && (
        <div className="fixed top-6 left-1/2 -translate-x-1/2 z-50 rounded-md bg-emerald-600 px-4 py-2 text-sm text-white shadow-lg">
          ✅ {mensajeFlash}
        </div>
      )}

      {accesoEditando && (
        <ModalEditarAcceso
          acceso={accesoEditando}
          sistemas={sistemas}
          alCerrar={() => setAccesoEditando(null)}
          alExito={() => { setAccesoEditando(null); void cargar(); }}
        />
      )}
      {accesoViendo && (
        <ModalVerDetalles
          acceso={accesoViendo}
          sistema={sistemasPorId.get(accesoViendo.sistema_destino_id)}
          alCerrar={() => setAccesoViendo(null)}
          alCopiar={copiar}
        />
      )}

      {/* Vista mobile: cards apilados */}
      <div className="md:hidden space-y-3">
        {cargando && <div className="text-center text-stone-500 py-6">Cargando…</div>}
        {!cargando && accesosFiltrados.length === 0 && (
          <div className="text-center text-stone-500 py-6 border border-stone-200 rounded-md">
            No hay accesos. Toca <strong>+ Registrar acceso</strong>.
          </div>
        )}
        {!cargando && accesosFiltrados.map((a) => {
          const s = sistemasPorId.get(a.sistema_destino_id);
          const inactivo = a.estado === "REVOCADO";
          const pwdVisible = !!pwdsVisibles[a.id];
          const pwd = pwdsCache[a.id];
          return (
            <div key={a.id} className={`rounded-md border border-stone-200 bg-white p-3 space-y-2 ${inactivo ? "opacity-60" : ""}`}>
              <div>
                <div className="flex items-start justify-between gap-2">
                  <button
                    type="button"
                    onClick={() => void abrirEnPestana(a)}
                    className="font-medium text-cocina-marron hover:underline text-left flex-1"
                    disabled={inactivo}
                  >
                    {a.titulo || s?.nombre || "(sin título)"} ↗
                  </button>
                  {inactivo && <span className="text-[10px] bg-stone-300 text-stone-700 px-2 py-0.5 rounded shrink-0">inactivo</span>}
                </div>
                <div className="text-xs text-stone-400 break-all mt-0.5">{s?.url_acceso ?? "—"}</div>
              </div>
              <div className="grid grid-cols-[60px_1fr] gap-2 text-sm">
                <span className="text-stone-500 text-xs">Usuario</span>
                <button
                  type="button"
                  onClick={() => void copiar(a.usuario_externo, "Usuario")}
                  className="text-left break-all hover:bg-stone-100 rounded px-1 py-0.5"
                  title="Tocar para copiar"
                >
                  {a.usuario_externo}
                </button>
                <span className="text-stone-500 text-xs">Clave</span>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => void copiarPwd(a)}
                    className="font-mono text-xs hover:bg-stone-100 rounded px-1 py-0.5 flex-1 text-left break-all"
                    title="Tocar para copiar"
                  >
                    {pwdVisible && pwd ? pwd : "••••••••"}
                  </button>
                  <button
                    type="button"
                    onClick={() => void togglePwdVisible(a)}
                    className="text-stone-500 shrink-0"
                  >{pwdVisible ? "🙈" : "👁️"}</button>
                </div>
              </div>
              {a.observaciones && (
                <div
                  className="text-xs text-stone-400 line-clamp-2 break-words"
                  title={a.observaciones}
                >{a.observaciones}</div>
              )}
              <div className="flex justify-end gap-4 pt-1 border-t border-stone-100">
                <button onClick={() => setAccesoViendo(a)} className="text-xs text-cocina-marron">Detalles</button>
                <button onClick={() => setAccesoEditando(a)} className="text-xs text-cocina-marron">Editar</button>
                {inactivo
                  ? <button onClick={() => void reactivar(a)} className="text-xs text-emerald-700">Reactivar</button>
                  : <button onClick={() => void desactivar(a)} className="text-xs text-red-700">Desactivar</button>
                }
              </div>
            </div>
          );
        })}
      </div>

      {/* Vista desktop: tabla */}
      <div className="hidden md:block overflow-x-auto rounded-md border border-stone-200">
        <table className="min-w-full text-sm">
          <thead className="bg-stone-100 text-left text-xs uppercase tracking-wider text-stone-600">
            <tr>
              <th className="w-12 px-3 py-2">#</th>
              <th className="px-3 py-2">Título / URL</th>
              <th className="px-3 py-2">Usuario</th>
              <th className="px-3 py-2">Contraseña</th>
              <th className="px-3 py-2 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody>
            {cargando && (
              <tr><td colSpan={5} className="px-3 py-6 text-center text-stone-500">Cargando…</td></tr>
            )}
            {!cargando && accesosFiltrados.length === 0 && (
              <tr><td colSpan={5} className="px-3 py-6 text-center text-stone-500">
                No hay accesos. Click en <strong>+ Registrar acceso</strong>.
              </td></tr>
            )}
            {!cargando && accesosFiltrados.map((a, idx) => {
              const s = sistemasPorId.get(a.sistema_destino_id);
              const inactivo = a.estado === "REVOCADO";
              const pwdVisible = !!pwdsVisibles[a.id];
              const pwd = pwdsCache[a.id];
              return (
                <tr key={a.id} className={`border-t border-stone-100 hover:bg-cocina-fondo align-top ${inactivo ? "opacity-60" : ""}`}>
                  <td className="px-3 py-2 text-stone-500">{idx + 1}</td>
                  <td className="px-3 py-2">
                    <button
                      type="button"
                      onClick={() => void abrirEnPestana(a)}
                      className="font-medium text-cocina-marron hover:underline text-left"
                      title="Abre el sistema y copia la clave al portapapeles"
                      disabled={inactivo}
                    >
                      {a.titulo || s?.nombre || "(sin título)"} ↗
                    </button>
                    {inactivo && <span className="ml-2 text-[10px] bg-stone-300 text-stone-700 px-2 py-0.5 rounded">inactivo</span>}
                    <div className="text-xs text-stone-400 break-all">{s?.url_acceso ?? "—"}</div>
                  </td>
                  <td className="px-3 py-2 text-cocina-oscuro max-w-[240px]">
                    <button
                      type="button"
                      onClick={() => void copiar(a.usuario_externo, "Usuario")}
                      className="text-left break-all hover:bg-stone-100 rounded px-1 py-0.5"
                      title="Click para copiar"
                    >
                      {a.usuario_externo}
                    </button>
                    {a.observaciones && (
                      <div
                        className="text-xs text-stone-400 line-clamp-2 break-words"
                        title={a.observaciones}
                      >{a.observaciones}</div>
                    )}
                  </td>
                  <td className="px-3 py-2 text-cocina-oscuro">
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => void copiarPwd(a)}
                        className="font-mono text-xs hover:bg-stone-100 rounded px-1 py-0.5 flex-1 text-left break-all"
                        title="Click para copiar"
                      >
                        {pwdVisible && pwd ? pwd : "••••••••"}
                      </button>
                      <button
                        type="button"
                        onClick={() => void togglePwdVisible(a)}
                        className="text-stone-500 hover:text-cocina-marron shrink-0"
                        title={pwdVisible ? "Ocultar" : "Ver"}
                      >{pwdVisible ? "🙈" : "👁️"}</button>
                    </div>
                  </td>
                  <td className="px-3 py-2 text-right whitespace-nowrap">
                    <div className="flex flex-wrap justify-end items-center gap-x-3 gap-y-1">
                      <button onClick={() => setAccesoViendo(a)} className="text-xs text-cocina-marron hover:underline">Detalles</button>
                      <button onClick={() => setAccesoEditando(a)} className="text-xs text-cocina-marron hover:underline">Editar</button>
                      {inactivo
                        ? <button onClick={() => void reactivar(a)} className="text-xs text-emerald-700 hover:underline">Reactivar</button>
                        : <button onClick={() => void desactivar(a)} className="text-xs text-red-700 hover:underline">Desactivar</button>
                      }
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

function ModalVerDetalles({ acceso, sistema, alCerrar, alCopiar }: {
  acceso: AccesoGuardado;
  sistema?: SistemaDisponible;
  alCerrar: () => void;
  alCopiar: (valor: string, etiqueta: string) => Promise<void>;
}) {
  const [credencial, setCredencial] = useState<CredencialAutofill | null>(null);
  const [pwdVisible, setPwdVisible] = useState(false);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  return (
    <ModalContenedor titulo="Detalles del acceso" alCerrar={alCerrar}>
      <dl className="space-y-3 text-sm">
        <Fila etiqueta="Título" valor={acceso.titulo || "—"} />
        <Fila etiqueta="Sistema" valor={sistema?.nombre || "—"} />
        <Fila etiqueta="URL" valor={sistema?.url_acceso || "—"} alCopiar={sistema?.url_acceso ? () => void alCopiar(sistema.url_acceso, "URL") : undefined} />
        <Fila etiqueta="Usuario" valor={acceso.usuario_externo} alCopiar={() => void alCopiar(acceso.usuario_externo, "Usuario")} />
        <div className="grid grid-cols-[90px_1fr_auto_auto] gap-2 items-start">
          <dt className="font-medium text-stone-500 text-xs uppercase tracking-wider pt-1">Clave</dt>
          <dd className="text-cocina-oscuro font-mono text-xs break-all">
            {cargando ? "Cargando…" : error ? <span className="text-red-700">{error}</span> :
              pwdVisible ? credencial?.password : "•".repeat(Math.min(20, credencial?.password.length ?? 8))}
          </dd>
          <button type="button" onClick={() => setPwdVisible(!pwdVisible)} className="text-stone-500 hover:text-cocina-marron" title={pwdVisible ? "Ocultar" : "Ver"}>{pwdVisible ? "🙈" : "👁️"}</button>
          <button type="button" onClick={() => credencial && void alCopiar(credencial.password, "Clave")} className="text-stone-500 hover:text-cocina-marron" title="Copiar">📋</button>
        </div>
        <Fila etiqueta="Tipo" valor={acceso.tipo || "WEB"} />
        {acceso.puerto != null && <Fila etiqueta="Puerto" valor={String(acceso.puerto)} />}
        <Fila etiqueta="Estado" valor={acceso.estado} />
        {acceso.observaciones && (
          <div className="grid grid-cols-[90px_1fr] gap-2 items-start">
            <dt className="font-medium text-stone-500 text-xs uppercase tracking-wider pt-1">Notas</dt>
            <dd className="text-cocina-oscuro text-xs whitespace-pre-wrap break-words">{acceso.observaciones}</dd>
          </div>
        )}
      </dl>
      <div className="mt-5 flex justify-end">
        <button type="button" onClick={alCerrar} className="boton-secundario">Cerrar</button>
      </div>
    </ModalContenedor>
  );
}

function Fila({ etiqueta, valor, alCopiar }: { etiqueta: string; valor: string; alCopiar?: () => void }) {
  return (
    <div className="grid grid-cols-[90px_1fr_auto] gap-2 items-start">
      <dt className="font-medium text-stone-500 text-xs uppercase tracking-wider pt-1">{etiqueta}</dt>
      <dd className="text-cocina-oscuro text-xs break-words">{valor}</dd>
      {alCopiar && <button onClick={alCopiar} className="text-stone-500 hover:text-cocina-marron" title="Copiar">📋</button>}
    </div>
  );
}

function ModalEditarAcceso({ acceso, sistemas, alCerrar, alExito }: { acceso: AccesoGuardado; sistemas: SistemaDisponible[]; alCerrar: () => void; alExito: () => void }) {
  const [titulo, setTitulo] = useState(acceso.titulo);
  const [sistemaId, setSistemaId] = useState(acceso.sistema_destino_id);
  const [usuario, setUsuario] = useState(acceso.usuario_externo);
  const [claveNueva, setClaveNueva] = useState("");
  const [tipo, setTipo] = useState<TipoAcceso>(acceso.tipo || "WEB");
  const [puerto, setPuerto] = useState<string>(acceso.puerto != null ? String(acceso.puerto) : "");
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
    if (puerto.trim()) {
      const n = Number(puerto);
      if (!Number.isInteger(n) || n < 1 || n > 65535) {
        setError("Puerto debe ser un entero entre 1 y 65535");
        return;
      }
    }
    setEnviando(true);
    try {
      const cuerpo: Record<string, unknown> = {
        titulo: titulo.trim(),
        sistema_destino_id: sistemaId,
        usuario_externo: usuario.trim(),
        observaciones: observaciones.trim(),
        tipo,
      };
      if (claveNueva) cuerpo.password = claveNueva;
      if (puerto.trim()) cuerpo.puerto = Number(puerto);
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
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label htmlFor="edit-puerto" className="etiqueta-campo">Puerto <span className="text-stone-400">(opcional)</span></label>
            <input id="edit-puerto" type="number" min={1} max={65535} className="campo-texto" value={puerto} onChange={(e) => setPuerto(e.target.value)} placeholder="opcional" />
          </div>
          <div>
            <label htmlFor="edit-tipo" className="etiqueta-campo">Tipo</label>
            <select id="edit-tipo" className="campo-texto" value={tipo} onChange={(e) => setTipo(e.target.value as TipoAcceso)}>
              {TIPOS_ACCESO.map((t) => (
                <option key={t} value={t}>{ETIQUETAS_TIPO[t]}</option>
              ))}
            </select>
          </div>
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
