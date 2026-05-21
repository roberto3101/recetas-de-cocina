import { useCallback, useEffect, useMemo, useState, FormEvent } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import CampoContrasena from "@/plataforma/ui/CampoContrasena";
import MedidorPassword, { evaluarPassword } from "@/plataforma/ui/MedidorPassword";
import SelectorSistemaConBuscador from "@/plataforma/ui/SelectorSistemaConBuscador";
import {
  AccesoGuardado,
  ETIQUETAS_TIPO,
  ListadoAccesos,
  ListadoSistemas,
  OPCIONES_ORDEN,
  SistemaDisponible,
  TAMANO_PAGINA_DEFAULT,
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
  // textoDebounced: lo que efectivamente se manda al backend. El debounce de
  // 250ms evita disparar una request por cada tecla. Mejora UX y evita
  // martillar la BD con queries inútiles mientras el usuario tipea.
  const [textoDebounced, setTextoDebounced] = useState("");
  const [sistemaIdFiltro, setSistemaIdFiltro] = useState<string>("");
  const [filtroEstado, setFiltroEstado] = useState<FiltroEstado>("ACTIVO");
  // ordenIndice indexa OPCIONES_ORDEN: 0=Más recientes, 1=Más antiguos, 2=A–Z, 3=Z–A
  const [ordenIndice, setOrdenIndice] = useState<number>(0);
  // Paginación server-side: paginaActual base 1 para UI, se traduce a offset al fetch
  const [paginaActual, setPaginaActual] = useState<number>(1);
  const [totalServer, setTotalServer] = useState<number>(0);

  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [accesos, setAccesos] = useState<AccesoGuardado[]>([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [pwdsVisibles, setPwdsVisibles] = useState<Record<string, boolean>>({});
  const [pwdsCache, setPwdsCache] = useState<Record<string, string>>({});
  const [mensajeFlash, setMensajeFlash] = useState<string | null>(null);

  const [accesoEditando, setAccesoEditando] = useState<AccesoGuardado | null>(null);
  const [accesoViendo, setAccesoViendo] = useState<AccesoGuardado | null>(null);
  const [accesoEliminando, setAccesoEliminando] = useState<AccesoGuardado | null>(null);
  const [mostrarExportar, setMostrarExportar] = useState(false);

  // Debounce del texto de búsqueda: el usuario tipea → esperamos 250ms sin
  // teclas nuevas → recién mandamos al backend. Cancelar/reiniciar con cada
  // tecla evita N requests intermedias.
  useEffect(() => {
    const id = setTimeout(() => setTextoDebounced(texto.trim()), 250);
    return () => clearTimeout(id);
  }, [texto]);

  const cargar = useCallback(async () => {
    setCargando(true);
    setError(null);
    try {
      const orden = OPCIONES_ORDEN[ordenIndice];
      const offset = (paginaActual - 1) * TAMANO_PAGINA_DEFAULT;
      const params = new URLSearchParams({
        limite: String(TAMANO_PAGINA_DEFAULT),
        offset: String(offset),
        orden: orden.campo,
        direccion: orden.direccion,
      });
      if (filtroEstado !== "TODOS") {
        params.set("estado", filtroEstado);
      }
      if (sistemaIdFiltro) {
        params.set("sistema_id", sistemaIdFiltro);
      }
      if (textoDebounced) {
        params.set("q", textoDebounced);
      }
      const [sis, acc] = await Promise.all([
        pedirJson<ListadoSistemas>("/cocina/sistemas"),
        pedirJson<ListadoAccesos>(`/cocina/boveda/accesos?${params}`),
      ]);
      setSistemas(sis.sistemas);
      setAccesos(acc.accesos);
      setTotalServer(acc.total);
    } catch (e) {
      if (e instanceof Error) setError(e.message);
    } finally {
      setCargando(false);
    }
  }, [ordenIndice, paginaActual, filtroEstado, sistemaIdFiltro, textoDebounced]);

  useEffect(() => { void cargar(); }, [cargar]);

  // Resetear a página 1 cuando cambian filtros que afectan el set de
  // resultados. Si no hacemos esto, el usuario puede quedar "atrapado" en
  // una página inexistente (ej: estaba en pág 3 de ACTIVO y al cambiar a
  // REVOCADO solo hay 1 página, o al filtrar por texto se reduce el total).
  useEffect(() => {
    setPaginaActual(1);
  }, [ordenIndice, filtroEstado, sistemaIdFiltro, textoDebounced]);

  const totalPaginas = Math.max(1, Math.ceil(totalServer / TAMANO_PAGINA_DEFAULT));

  const sistemasPorId = useMemo(() => {
    const m = new Map<string, SistemaDisponible>();
    for (const s of sistemas) m.set(s.id, s);
    return m;
  }, [sistemas]);

  // Ya NO filtramos client-side. Todo lo que llega en `accesos` viene
  // pre-filtrado por el backend con q + sistema_id + estado. Esto garantiza
  // que la búsqueda funcione sobre las 10K filas y no solo sobre la página
  // visible (bug previo).
  //
  // Limitación conocida: la búsqueda por `q` matchea contra titulo, nombre
  // del sistema y url_acceso del sistema. NO matchea contra usuario_externo
  // porque ese campo vive cifrado con AES-GCM en BD — sin acceso a la clave
  // no se puede hacer substring search. Se acepta como trade-off de seguridad.
  const accesosFiltrados = accesos;

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

  // Llamado desde el modal cuando el usuario tipea "ELIMINAR" y confirma.
  // Hard delete (no recuperable). El backend exige 2FA + audita la acción
  // antes del DELETE para que quede trazabilidad aunque la fila ya no exista.
  async function eliminarPermanente(a: AccesoGuardado) {
    try {
      await enviarJson(`/cocina/boveda/accesos/${a.id}/permanente`, undefined, "DELETE");
      setAccesoEliminando(null);
      flash("Acceso eliminado permanentemente");
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
            {cargando
              ? "Cargando…"
              : totalServer === 0
                ? texto || sistemaIdFiltro || filtroEstado !== "ACTIVO"
                  ? "Sin resultados para los filtros actuales"
                  : "Sin accesos"
                : totalPaginas > 1
                  ? `${totalServer} accesos · página ${paginaActual} de ${totalPaginas}`
                  : `${totalServer} ${totalServer === 1 ? "acceso" : "accesos"}`}
          </p>
        </div>
        <div className="flex gap-2 flex-wrap">
          <button onClick={() => setMostrarExportar(true)} className="boton-secundario" title="Descargar todos los accesos en un ZIP cifrado">
            ⬇ Exportar
          </button>
          <button onClick={() => navegar("/panel/registro")} className="boton-primario">+ Registrar acceso</button>
        </div>
      </div>

      <form className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4" onSubmit={(e) => e.preventDefault()}>
        <input
          type="text"
          placeholder="Buscar por título o sistema…"
          className="campo-texto"
          value={texto}
          onChange={(e) => setTexto(e.target.value)}
          title="Busca en título del acceso, nombre del sistema y URL. El usuario externo NO se puede buscar (está cifrado)."
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
        <select
          className="campo-texto"
          value={ordenIndice}
          onChange={(e) => setOrdenIndice(Number(e.target.value))}
          aria-label="Ordenar por"
        >
          {OPCIONES_ORDEN.map((o, i) => (
            <option key={`${o.campo}-${o.direccion}`} value={i}>
              {o.etiqueta}
            </option>
          ))}
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
      {accesoEliminando && (
        <ModalConfirmarEliminar
          acceso={accesoEliminando}
          sistema={sistemasPorId.get(accesoEliminando.sistema_destino_id)}
          alCerrar={() => setAccesoEliminando(null)}
          alConfirmar={() => void eliminarPermanente(accesoEliminando)}
        />
      )}
      {mostrarExportar && (
        <ModalExportarBoveda
          totalAccesos={totalServer}
          alCerrar={() => setMostrarExportar(false)}
        />
      )}

      {/* Vista mobile: cards apilados.
          Para evitar saltos de scroll al cambiar de página, NO se borran
          las filas durante un refetch. Solo se atenúa la opacidad mientras
          carga. La layout total se mantiene → la barra de paginación
          (al fondo) no se mueve y el usuario queda en el mismo punto. */}
      <div className={`md:hidden space-y-3 transition-opacity ${cargando && accesosFiltrados.length > 0 ? "opacity-60" : ""}`}>
        {cargando && accesosFiltrados.length === 0 && (
          <div className="text-center text-stone-500 py-6">Cargando…</div>
        )}
        {!cargando && accesosFiltrados.length === 0 && (
          <div className="text-center text-stone-500 py-6 border border-stone-200 rounded-md">
            No hay accesos. Toca <strong>+ Registrar acceso</strong>.
          </div>
        )}
        {accesosFiltrados.map((a) => {
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
              <div className="flex flex-wrap justify-end gap-x-4 gap-y-1 pt-1 border-t border-stone-100">
                <button onClick={() => setAccesoViendo(a)} className="text-xs text-cocina-marron">Detalles</button>
                <button onClick={() => setAccesoEditando(a)} className="text-xs text-cocina-marron">Editar</button>
                {inactivo
                  ? <button onClick={() => void reactivar(a)} className="text-xs text-emerald-700">Reactivar</button>
                  : <button onClick={() => void desactivar(a)} className="text-xs text-red-700">Desactivar</button>
                }
                <button onClick={() => setAccesoEliminando(a)} className="text-xs text-red-800 font-medium">Eliminar</button>
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
          <tbody className={cargando && accesosFiltrados.length > 0 ? "opacity-60 transition-opacity" : ""}>
            {cargando && accesosFiltrados.length === 0 && (
              <tr><td colSpan={5} className="px-3 py-6 text-center text-stone-500">Cargando…</td></tr>
            )}
            {!cargando && accesosFiltrados.length === 0 && (
              <tr><td colSpan={5} className="px-3 py-6 text-center text-stone-500">
                No hay accesos. Click en <strong>+ Registrar acceso</strong>.
              </td></tr>
            )}
            {accesosFiltrados.map((a, idx) => {
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
                      <button onClick={() => setAccesoEliminando(a)} className="text-xs text-red-800 font-medium hover:underline">Eliminar</button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Paginación: solo se muestra si el total excede el tamaño de página */}
      {totalPaginas > 1 && (
        <div className="flex items-center justify-between gap-3 border-t border-stone-200 pt-3 text-sm">
          <span className="text-stone-500">
            Página <span className="font-medium text-cocina-oscuro">{paginaActual}</span> de{" "}
            <span className="font-medium text-cocina-oscuro">{totalPaginas}</span>
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              className="boton-secundario disabled:opacity-40"
              onClick={() => setPaginaActual((p) => Math.max(1, p - 1))}
              disabled={paginaActual <= 1 || cargando}
            >
              ← Anterior
            </button>
            <button
              type="button"
              className="boton-secundario disabled:opacity-40"
              onClick={() => setPaginaActual((p) => Math.min(totalPaginas, p + 1))}
              disabled={paginaActual >= totalPaginas || cargando}
            >
              Siguiente →
            </button>
          </div>
        </div>
      )}
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
        {/* Clave: en móvil etiqueta arriba, valor + botones abajo; en sm+ todo en línea */}
        <div className="grid grid-cols-1 sm:grid-cols-[90px_minmax(0,1fr)_auto_auto] gap-1 sm:gap-2 items-start">
          <dt className="font-medium text-stone-500 text-xs uppercase tracking-wider pt-1">Clave</dt>
          <dd className="min-w-0 text-cocina-oscuro font-mono text-xs break-all">
            {cargando ? "Cargando…" : error ? <span className="text-red-700">{error}</span> :
              pwdVisible ? credencial?.password : "•".repeat(Math.min(20, credencial?.password.length ?? 8))}
          </dd>
          <div className="col-span-1 sm:col-auto flex gap-2 sm:gap-0">
            <button type="button" onClick={() => setPwdVisible(!pwdVisible)} className="text-stone-500 hover:text-cocina-marron sm:px-1" title={pwdVisible ? "Ocultar" : "Ver"}>{pwdVisible ? "🙈" : "👁️"}</button>
            <button type="button" onClick={() => credencial && void alCopiar(credencial.password, "Clave")} className="text-stone-500 hover:text-cocina-marron sm:px-1" title="Copiar">📋</button>
          </div>
        </div>
        <Fila etiqueta="Tipo" valor={acceso.tipo || "WEB"} />
        {acceso.puerto != null && <Fila etiqueta="Puerto" valor={String(acceso.puerto)} />}
        <Fila etiqueta="Estado" valor={acceso.estado} />
        {acceso.observaciones && (
          <div className="grid grid-cols-1 sm:grid-cols-[90px_minmax(0,1fr)] gap-1 sm:gap-2 items-start">
            <dt className="font-medium text-stone-500 text-xs uppercase tracking-wider pt-1">Notas</dt>
            <dd className="min-w-0 max-h-48 overflow-y-auto rounded border border-stone-100 bg-stone-50 px-2 py-1 text-cocina-oscuro text-xs whitespace-pre-wrap break-words">{acceso.observaciones}</dd>
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
    <div className="grid grid-cols-1 sm:grid-cols-[90px_minmax(0,1fr)_auto] gap-1 sm:gap-2 items-start">
      <dt className="font-medium text-stone-500 text-xs uppercase tracking-wider pt-1">{etiqueta}</dt>
      <dd className="min-w-0 text-cocina-oscuro text-xs break-words">{valor}</dd>
      {alCopiar && <button onClick={alCopiar} className="text-stone-500 hover:text-cocina-marron justify-self-start sm:justify-self-auto" title="Copiar">📋</button>}
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
          <SelectorSistemaConBuscador
            idCampo="edit-sistema"
            sistemas={sistemas}
            valor={sistemaId}
            alCambiar={(nuevoId) => setSistemaId(nuevoId)}
          />
        </div>
        <div>
          <label htmlFor="edit-usuario" className="etiqueta-campo">Usuario</label>
          <input id="edit-usuario" type="text" className="campo-texto" value={usuario} onChange={(e) => setUsuario(e.target.value)} required maxLength={200} />
        </div>
        <div>
          <label htmlFor="edit-clave" className="etiqueta-campo">Nueva clave <span className="text-stone-400">(opcional, deja vacío para mantener la actual)</span></label>
          <CampoContrasena
            id="edit-clave"
            value={claveNueva}
            onChange={(e) => setClaveNueva(e.target.value)}
            autoComplete="new-password"
            maxLength={500}
          />
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

// ModalConfirmarEliminar: confirmación type-to-confirm para hard delete.
// El usuario debe escribir literalmente "ELIMINAR" (case-sensitive) en el
// input para habilitar el botón de confirmación. Esto previene clicks
// accidentales o ataques de click-jacking que disparen el borrado.
//
// Responsive:
//   - max-w-md (32rem) en el contenedor — entra cómodo en mobile (paddding p-4
//     en el overlay deja espacio para que no se pegue a los bordes).
//   - Botones se apilan vertical en xs y horizontal en sm+.
//   - El input ocupa todo el ancho disponible.
function ModalConfirmarEliminar({
  acceso,
  sistema,
  alCerrar,
  alConfirmar,
}: {
  acceso: AccesoGuardado;
  sistema?: SistemaDisponible;
  alCerrar: () => void;
  alConfirmar: () => void;
}) {
  const [valor, setValor] = useState("");
  const [enviando, setEnviando] = useState(false);
  const palabraClave = "ELIMINAR";
  const coincide = valor === palabraClave;

  async function alSubmit(evento: FormEvent) {
    evento.preventDefault();
    if (!coincide || enviando) return;
    setEnviando(true);
    try {
      alConfirmar();
    } finally {
      // alConfirmar es async pero no esperamos acá (lo dispara fire-and-forget
      // contra el backend). El padre se encarga de cerrar el modal al recibir
      // éxito. Si el padre no cierra, el usuario puede hacerlo manualmente.
      setEnviando(false);
    }
  }

  return (
    <ModalContenedor titulo="Eliminar acceso" alCerrar={alCerrar}>
      <form onSubmit={alSubmit} className="space-y-4">
        <div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-900">
          <p className="font-semibold mb-1">Esta acción NO se puede deshacer.</p>
          <p className="text-xs">
            Se borrará permanentemente la fila de la base de datos, incluyendo
            la contraseña cifrada. No hay manera de recuperarlo después.
            {" "}
            Si solo querés ocultarlo del listado, usá <strong>Desactivar</strong> en su lugar.
          </p>
        </div>

        <div className="text-sm">
          <p className="text-stone-500 text-xs uppercase tracking-wider mb-1">Vas a eliminar</p>
          <p className="font-medium text-cocina-oscuro break-words">
            {acceso.titulo || acceso.usuario_externo}
          </p>
          <p className="text-xs text-stone-500 break-words">
            {sistema?.nombre ?? "—"} · usuario: {acceso.usuario_externo}
          </p>
        </div>

        <div>
          <label htmlFor="confirmar-eliminar" className="etiqueta-campo">
            ¿Estás seguro que deseas eliminar esto? Escribí <span className="font-mono text-red-700">{palabraClave}</span> para confirmar.
          </label>
          <input
            id="confirmar-eliminar"
            type="text"
            autoFocus
            autoComplete="off"
            spellCheck={false}
            className={`campo-texto font-mono ${coincide ? "border-red-500 focus:ring-red-300" : ""}`}
            value={valor}
            onChange={(e) => setValor(e.target.value)}
            placeholder={palabraClave}
            aria-describedby="ayuda-confirmar"
          />
          <p id="ayuda-confirmar" className="mt-1 text-xs text-stone-500">
            Tiene que coincidir exactamente (en mayúsculas).
          </p>
        </div>

        <div className="flex flex-col-reverse sm:flex-row justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={alCerrar}
            className="boton-secundario"
            disabled={enviando}
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={!coincide || enviando}
            className="inline-flex items-center justify-center rounded-md bg-red-700 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-800 focus:outline-none focus:ring-2 focus:ring-red-700 focus:ring-offset-2 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {enviando ? "Eliminando…" : "Eliminar permanentemente"}
          </button>
        </div>
      </form>
    </ModalContenedor>
  );
}

// ModalExportarBoveda: descarga un ZIP cifrado AES-256 con todos los
// accesos descifrados. La passphrase la elige el usuario en este modal
// (no se reusa la del login). Si pierde la passphrase, el ZIP queda
// inservible — por eso lo aclaramos en el copy.
//
// Mecánica del download:
//   - fetch() devuelve Blob binario (no JSON)
//   - Creamos un object URL del blob
//   - Insertamos un <a download="..."> y simulamos click
//   - Limpiamos el object URL
// Es la receta estándar para descargar binarios desde un endpoint
// autenticado sin window.location.href (que perdería la sesión POST).
function ModalExportarBoveda({
  totalAccesos,
  alCerrar,
}: {
  totalAccesos: number;
  alCerrar: () => void;
}) {
  const [passphrase, setPassphrase] = useState("");
  const [confirmacion, setConfirmacion] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fortaleza = evaluarPassword(passphrase);
  const coincide = passphrase.length > 0 && passphrase === confirmacion;
  const puedeEnviar = fortaleza.todoOk && coincide && !enviando;

  async function alExportar(evento: FormEvent) {
    evento.preventDefault();
    if (!puedeEnviar) return;
    setError(null);
    setEnviando(true);
    try {
      const respuesta = await fetch("/cocina/boveda/accesos/exportar", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ passphrase }),
      });
      if (!respuesta.ok) {
        const tipo = respuesta.headers.get("Content-Type") ?? "";
        if (tipo.includes("application/json")) {
          const cuerpo = (await respuesta.json()) as { error?: string };
          throw new Error(cuerpo.error ?? "Error al exportar");
        }
        throw new Error(`Error HTTP ${respuesta.status}`);
      }
      const disp = respuesta.headers.get("Content-Disposition") ?? "";
      const match = disp.match(/filename="([^"]+)"/);
      const nombreArchivo = match?.[1] ?? `gestor-codeplex-${new Date().toISOString().slice(0, 10)}.zip`;

      const blob = await respuesta.blob();
      const url = URL.createObjectURL(blob);
      const ancla = document.createElement("a");
      ancla.href = url;
      ancla.download = nombreArchivo;
      document.body.appendChild(ancla);
      ancla.click();
      document.body.removeChild(ancla);
      URL.revokeObjectURL(url);

      alCerrar();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <ModalContenedor titulo="Exportar bóveda" alCerrar={alCerrar}>
      <form onSubmit={alExportar} className="space-y-4">
        <div className="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
          <p>
            Vas a descargar un <strong>ZIP cifrado con AES-256</strong> con los{" "}
            <strong>{totalAccesos} accesos</strong> en plano dentro. La passphrase
            la elegís acá y la necesitarás para abrirlo después.
          </p>
          <p className="mt-2 text-xs">
            <strong>Si perdés la passphrase, el ZIP queda inservible.</strong>{" "}
            Guardala en tu password manager personal. El ZIP en sí podés
            guardarlo donde quieras — sin la passphrase no se puede abrir.
          </p>
        </div>

        <div>
          <label className="etiqueta-campo">Passphrase para el ZIP</label>
          <CampoContrasena
            required
            value={passphrase}
            onChange={(e) => setPassphrase(e.target.value)}
            autoComplete="new-password"
            validez={passphrase.length === 0 ? undefined : fortaleza.todoOk ? "ok" : "error"}
          />
          <MedidorPassword valor={passphrase} confirmacion={confirmacion} />
        </div>

        <div>
          <label className="etiqueta-campo">Confirmar passphrase</label>
          <CampoContrasena
            required
            value={confirmacion}
            onChange={(e) => setConfirmacion(e.target.value)}
            autoComplete="new-password"
            validez={confirmacion.length === 0 ? undefined : coincide ? "ok" : "error"}
          />
        </div>

        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="flex flex-col-reverse sm:flex-row justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={alCerrar}
            className="boton-secundario"
            disabled={enviando}
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={!puedeEnviar}
            className="boton-primario disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {enviando ? "Generando ZIP…" : "Descargar ZIP cifrado"}
          </button>
        </div>
      </form>
    </ModalContenedor>
  );
}
