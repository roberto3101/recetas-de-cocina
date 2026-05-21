import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  ListadoSistemas,
  SistemaDisponible,
} from "@/capacidades/accesos/tipos";

// Tipo de respuesta del endpoint que devuelve cuántos accesos ACTIVOS
// hay vinculados a un sistema. Lo usamos antes de mostrar el modal de
// confirmación para alertar al usuario del efecto cascada.
type ConteoAccesos = { accesos_activos: number };
type ConteoRevocados = { accesos_revocados: number };
type ResultadoDesactivacion = { sistema_eliminado: boolean; accesos_revocados: number };
type ResultadoReactivacion = { sistema_reactivado: boolean; accesos_reactivados: number };

// Pestaña del listado: "activos" filtra estado=ACTIVO (default, lo que aparece
// en selectores); "archivados" filtra estado=ELIMINADO (lista el archivo, con
// botón Reactivar para sacarlos del archivo).
type Pestana = "activos" | "archivados";

function slugificar(s: string): string {
  return s
    .toLowerCase()
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .slice(0, 50);
}

export default function PaginaSistemas() {
  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mensaje, setMensaje] = useState<string | null>(null);
  const [mostrandoForm, setMostrandoForm] = useState(false);
  const [pestana, setPestana] = useState<Pestana>("activos");
  // Búsqueda client-side: el número total de sistemas es chico (decenas),
  // no vale la pena mandar query al backend. Filtramos en memoria sobre la
  // lista ya cargada del backend para la pestaña activa.
  const [busqueda, setBusqueda] = useState("");

  const [nombre, setNombre] = useState("");
  const [urlAcceso, setUrlAcceso] = useState("");
  const [enviando, setEnviando] = useState(false);

  const cargar = useCallback(async () => {
    setCargando(true);
    try {
      // En pestaña Archivados consultamos los ELIMINADOS. El backend
      // hace whitelist del param, valores raros caen al default ACTIVO.
      const url = pestana === "archivados"
        ? "/cocina/sistemas?estado=ELIMINADO"
        : "/cocina/sistemas";
      const r = await pedirJson<ListadoSistemas>(url);
      setSistemas(r.sistemas);
    } catch (e) {
      if (e instanceof Error) setError(e.message);
    } finally {
      setCargando(false);
    }
  }, [pestana]);

  useEffect(() => { void cargar(); }, [cargar]);

  // Al cambiar de pestaña, ocultamos el form de "Nuevo sistema" (solo aplica
  // en Activos), limpiamos búsqueda y mensajes para evitar confusión visual.
  useEffect(() => {
    setMostrandoForm(false);
    setMensaje(null);
    setError(null);
    setBusqueda("");
  }, [pestana]);

  // Filtrado case-insensitive sobre nombre, URL y código del sistema.
  // useMemo evita recalcular en cada render si nada cambió.
  const sistemasFiltrados = useMemo(() => {
    const q = busqueda.trim().toLowerCase();
    if (!q) return sistemas;
    return sistemas.filter((s) =>
      s.nombre.toLowerCase().includes(q) ||
      s.url_acceso.toLowerCase().includes(q) ||
      s.codigo.toLowerCase().includes(q),
    );
  }, [sistemas, busqueda]);

  function limpiarForm() {
    setNombre("");
    setUrlAcceso("");
  }

  async function alGuardar(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setMensaje(null);
    setEnviando(true);
    try {
      const codigoAuto = slugificar(nombre) || `sistema_${Date.now()}`;
      await enviarJson("/cocina/sistemas", {
        codigo: codigoAuto,
        nombre: nombre.trim(),
        url_acceso: urlAcceso.trim(),
        // resto: defaults sensatos. Si después necesitamos algo distinto, se edita.
        url_login: urlAcceso.trim(),
        nombre_campo_usuario: "",
        nombre_campo_password: "",
        metodo_login: "POST",
      });
      setMensaje(`Sistema "${nombre}" registrado.`);
      limpiarForm();
      setMostrandoForm(false);
      void cargar();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  // sistemaDesactivando: cuando hay un sistema acá, se renderiza el modal
  // type-to-confirm. Acompañado del count de accesos vinculados (consultado
  // al backend antes de mostrar el modal) para que el usuario sepa cuántos
  // se van a revocar en cascada.
  const [sistemaDesactivando, setSistemaDesactivando] = useState<SistemaDisponible | null>(null);
  const [accesosVinculados, setAccesosVinculados] = useState<number | null>(null);

  async function abrirModalDesactivar(s: SistemaDisponible) {
    setError(null);
    setMensaje(null);
    setAccesosVinculados(null);
    setSistemaDesactivando(s);
    try {
      // Preguntar cuántos accesos quedan vinculados ANTES de mostrar el modal.
      // Si falla la consulta, mostramos el modal igual con "?" y dejamos que
      // el usuario decida — no bloqueamos la acción por un GET que falló.
      const r = await pedirJson<ConteoAccesos>(`/cocina/sistemas/${s.id}/accesos-activos`);
      setAccesosVinculados(r.accesos_activos);
    } catch {
      setAccesosVinculados(null);
    }
  }

  async function confirmarDesactivacion(s: SistemaDisponible) {
    try {
      const r = await enviarJson<ResultadoDesactivacion>(`/cocina/sistemas/${s.id}`, undefined, "DELETE");
      setSistemaDesactivando(null);
      if (r.accesos_revocados > 0) {
        setMensaje(`Sistema "${s.nombre}" desactivado. ${r.accesos_revocados} acceso(s) vinculado(s) fueron revocados en cascada.`);
      } else {
        setMensaje(`Sistema "${s.nombre}" desactivado.`);
      }
      void cargar();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    }
  }

  // Reactivar es la operación inversa de desactivar. Mismo patrón:
  // 1) Modal con confirmación que muestra cuántos accesos van a revivir
  // 2) POST al backend, que ejecuta la cascada en una transacción atómica
  const [sistemaReactivando, setSistemaReactivando] = useState<SistemaDisponible | null>(null);
  const [accesosRevocados, setAccesosRevocados] = useState<number | null>(null);

  async function abrirModalReactivar(s: SistemaDisponible) {
    setError(null);
    setMensaje(null);
    setAccesosRevocados(null);
    setSistemaReactivando(s);
    try {
      const r = await pedirJson<ConteoRevocados>(`/cocina/sistemas/${s.id}/accesos-revocados`);
      setAccesosRevocados(r.accesos_revocados);
    } catch {
      setAccesosRevocados(null);
    }
  }

  async function confirmarReactivacion(s: SistemaDisponible) {
    try {
      const r = await enviarJson<ResultadoReactivacion>(`/cocina/sistemas/${s.id}/reactivar`, undefined, "POST");
      setSistemaReactivando(null);
      if (r.accesos_reactivados > 0) {
        setMensaje(`Sistema "${s.nombre}" reactivado. ${r.accesos_reactivados} acceso(s) vinculado(s) volvieron a estar activos.`);
      } else {
        setMensaje(`Sistema "${s.nombre}" reactivado.`);
      }
      void cargar();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4 flex-wrap">
        <div>
          <h2 className="text-2xl font-semibold text-cocina-oscuro">Sistemas</h2>
          <p className="text-sm text-stone-500">
            {pestana === "activos"
              ? <>Catálogo de URLs a las que vas a guardar credenciales. {sistemas.length} sistema{sistemas.length === 1 ? "" : "s"} activo{sistemas.length === 1 ? "" : "s"}.</>
              : <>Sistemas archivados. {sistemas.length} en el archivo. Reactivá uno para volver a usarlo y revivir sus accesos.</>}
          </p>
        </div>
        {pestana === "activos" && (
          <button
            className="boton-primario"
            onClick={() => { setMostrandoForm(!mostrandoForm); setMensaje(null); setError(null); }}
          >
            {mostrandoForm ? "Cancelar" : "+ Nuevo sistema"}
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-stone-200">
        <button
          type="button"
          onClick={() => setPestana("activos")}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            pestana === "activos"
              ? "border-cocina-marron text-cocina-marron"
              : "border-transparent text-stone-500 hover:text-cocina-oscuro"
          }`}
        >
          Activos
        </button>
        <button
          type="button"
          onClick={() => setPestana("archivados")}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            pestana === "archivados"
              ? "border-cocina-marron text-cocina-marron"
              : "border-transparent text-stone-500 hover:text-cocina-oscuro"
          }`}
        >
          Archivados
        </button>
      </div>

      {/* Buscador. Aparece siempre que haya al menos un sistema cargado;
          si la lista está vacía no tiene sentido un buscador. */}
      {sistemas.length > 0 && (
        <input
          type="text"
          placeholder="Buscar por nombre, URL o código…"
          className="campo-texto"
          value={busqueda}
          onChange={(e) => setBusqueda(e.target.value)}
        />
      )}

      {mensaje && <div className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-2 text-sm text-emerald-800">{mensaje}</div>}
      {error && <div className="rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">{error}</div>}

      {mostrandoForm && (
        <form onSubmit={alGuardar} className="rounded-md border border-stone-200 bg-white p-4 space-y-3">
          <div>
            <label htmlFor="sis-nombre" className="etiqueta-campo">Nombre</label>
            <input
              id="sis-nombre"
              required
              type="text"
              className="campo-texto"
              value={nombre}
              onChange={(e) => setNombre(e.target.value)}
              placeholder="CRM Codeplex"
              autoFocus
            />
          </div>
          <div>
            <label htmlFor="sis-url" className="etiqueta-campo">URL</label>
            <input
              id="sis-url"
              required
              type="url"
              className="campo-texto"
              value={urlAcceso}
              onChange={(e) => setUrlAcceso(e.target.value)}
              placeholder="https://codeplex.pe/crm/admin/authentication"
            />
            <p className="text-xs text-stone-400 mt-1">Pega la URL del login del sistema, tal cual.</p>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" className="boton-secundario" onClick={() => setMostrandoForm(false)} disabled={enviando}>Cancelar</button>
            <button type="submit" className="boton-primario" disabled={enviando}>
              {enviando ? "Guardando…" : "Guardar"}
            </button>
          </div>
        </form>
      )}

      {/* Mobile: cards */}
      <div className="md:hidden space-y-3">
        {cargando && <div className="text-center text-stone-500 py-6">Cargando…</div>}
        {!cargando && sistemas.length === 0 && (
          <div className="text-center text-stone-500 py-6 border border-stone-200 rounded-md">
            {pestana === "activos"
              ? <>Sin sistemas activos. Toca <strong>+ Nuevo sistema</strong>.</>
              : "Sin sistemas archivados."}
          </div>
        )}
        {!cargando && sistemas.length > 0 && sistemasFiltrados.length === 0 && (
          <div className="text-center text-stone-500 py-6 border border-stone-200 rounded-md">
            Sin resultados para "{busqueda}".
          </div>
        )}
        {!cargando && sistemasFiltrados.map((s) => (
          <div key={s.id} className="rounded-md border border-stone-200 bg-white p-3 space-y-1">
            <div className="font-medium text-cocina-oscuro">{s.nombre}</div>
            <div className="text-cocina-marron break-all text-xs">{s.url_acceso}</div>
            <div className="flex justify-end pt-1 border-t border-stone-100">
              {pestana === "activos"
                ? <button onClick={() => void abrirModalDesactivar(s)} className="text-xs text-red-700">Desactivar</button>
                : <button onClick={() => void abrirModalReactivar(s)} className="text-xs text-emerald-700">Reactivar</button>}
            </div>
          </div>
        ))}
      </div>

      {/* Desktop: tabla */}
      <div className="hidden md:block overflow-x-auto rounded-md border border-stone-200">
        <table className="min-w-full text-sm">
          <thead className="bg-stone-100 text-left text-xs uppercase tracking-wider text-stone-600">
            <tr>
              <th className="px-3 py-2">Nombre</th>
              <th className="px-3 py-2">URL</th>
              <th className="px-3 py-2 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody>
            {cargando && (<tr><td colSpan={3} className="px-3 py-6 text-center text-stone-500">Cargando…</td></tr>)}
            {!cargando && sistemas.length === 0 && (
              <tr><td colSpan={3} className="px-3 py-6 text-center text-stone-500">
                {pestana === "activos"
                  ? <>Sin sistemas activos. Click en <strong>+ Nuevo sistema</strong>.</>
                  : "Sin sistemas archivados."}
              </td></tr>
            )}
            {!cargando && sistemas.length > 0 && sistemasFiltrados.length === 0 && (
              <tr><td colSpan={3} className="px-3 py-6 text-center text-stone-500">
                Sin resultados para "{busqueda}".
              </td></tr>
            )}
            {!cargando && sistemasFiltrados.map((s) => (
              <tr key={s.id} className="border-t border-stone-100 hover:bg-cocina-fondo">
                <td className="px-3 py-2 font-medium text-cocina-oscuro">{s.nombre}</td>
                <td className="px-3 py-2 text-cocina-marron break-all text-xs">{s.url_acceso}</td>
                <td className="px-3 py-2 text-right whitespace-nowrap">
                  {pestana === "activos"
                    ? <button onClick={() => void abrirModalDesactivar(s)} className="text-xs text-red-700 hover:underline">Desactivar</button>
                    : <button onClick={() => void abrirModalReactivar(s)} className="text-xs text-emerald-700 hover:underline">Reactivar</button>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {sistemaDesactivando && (
        <ModalDesactivarSistema
          sistema={sistemaDesactivando}
          accesosVinculados={accesosVinculados}
          alCerrar={() => setSistemaDesactivando(null)}
          alConfirmar={() => void confirmarDesactivacion(sistemaDesactivando)}
        />
      )}
      {sistemaReactivando && (
        <ModalReactivarSistema
          sistema={sistemaReactivando}
          accesosVinculados={accesosRevocados}
          alCerrar={() => setSistemaReactivando(null)}
          alConfirmar={() => void confirmarReactivacion(sistemaReactivando)}
        />
      )}
    </div>
  );
}

// ModalDesactivarSistema: type-to-confirm para soft-delete de un sistema.
// Muestra cuántos accesos vinculados se van a revocar EN CASCADA si el
// usuario confirma. Requiere escribir "DESACTIVAR" para habilitar el botón.
// Nada se borra físicamente — reversible si se reactiva el sistema.
function ModalDesactivarSistema({
  sistema,
  accesosVinculados,
  alCerrar,
  alConfirmar,
}: {
  sistema: SistemaDisponible;
  accesosVinculados: number | null;
  alCerrar: () => void;
  alConfirmar: () => void;
}) {
  const [valor, setValor] = useState("");
  const [enviando, setEnviando] = useState(false);
  const palabraClave = "DESACTIVAR";
  const coincide = valor === palabraClave;

  function alSubmit(e: FormEvent) {
    e.preventDefault();
    if (!coincide || enviando) return;
    setEnviando(true);
    alConfirmar();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={alCerrar}>
      <div className="w-full max-w-lg rounded-md bg-white shadow-lg" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-stone-200 px-5 py-3">
          <h3 className="text-lg font-semibold text-cocina-oscuro">Desactivar sistema</h3>
          <button type="button" onClick={alCerrar} className="text-stone-500 hover:text-cocina-marron" aria-label="Cerrar">✕</button>
        </div>

        <form onSubmit={alSubmit} className="px-5 py-4 space-y-4">
          <div className="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
            <p className="font-semibold mb-1">Soft delete con cascada.</p>
            <p className="text-xs">
              El sistema se marcará como <strong>ELIMINADO</strong> y desaparecerá del listado y del selector al
              registrar accesos. {" "}
              {accesosVinculados === null
                ? "Sus accesos vinculados también serán revocados."
                : accesosVinculados === 0
                  ? <>No tiene accesos vinculados — nada en cascada.</>
                  : <>Se revocarán en cascada <strong>{accesosVinculados} acceso{accesosVinculados === 1 ? "" : "s"}</strong> vinculado{accesosVinculados === 1 ? "" : "s"}.</>}
              {" "}Nada se borra físicamente — se puede reactivar manualmente en BD si te equivocas.
            </p>
          </div>

          <div className="text-sm">
            <p className="text-stone-500 text-xs uppercase tracking-wider mb-1">Vas a desactivar</p>
            <p className="font-medium text-cocina-oscuro break-words">{sistema.nombre}</p>
            <p className="text-xs text-stone-500 break-words">{sistema.url_acceso}</p>
          </div>

          <div>
            <label htmlFor="confirmar-desactivar" className="etiqueta-campo">
              Escribí <span className="font-mono text-red-700">{palabraClave}</span> para confirmar.
            </label>
            <input
              id="confirmar-desactivar"
              type="text"
              autoFocus
              autoComplete="off"
              spellCheck={false}
              className={`campo-texto font-mono ${coincide ? "border-red-500 focus:ring-red-300" : ""}`}
              value={valor}
              onChange={(e) => setValor(e.target.value)}
              placeholder={palabraClave}
            />
            <p className="mt-1 text-xs text-stone-500">Tiene que coincidir exactamente (en mayúsculas).</p>
          </div>

          <div className="flex flex-col-reverse sm:flex-row justify-end gap-2 pt-2">
            <button type="button" onClick={alCerrar} className="boton-secundario" disabled={enviando}>
              Cancelar
            </button>
            <button
              type="submit"
              disabled={!coincide || enviando}
              className="inline-flex items-center justify-center rounded-md bg-red-700 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-800 focus:outline-none focus:ring-2 focus:ring-red-700 focus:ring-offset-2 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {enviando ? "Desactivando…" : "Desactivar sistema"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ModalReactivarSistema: confirmación para sacar al sistema del archivo y
// revivir sus accesos en cascada. UX más liviana que la de desactivar
// (botón normal, no type-to-confirm) porque es una operación constructiva,
// reversible al desactivarlo de nuevo si el operador se arrepiente.
function ModalReactivarSistema({
  sistema,
  accesosVinculados,
  alCerrar,
  alConfirmar,
}: {
  sistema: SistemaDisponible;
  accesosVinculados: number | null;
  alCerrar: () => void;
  alConfirmar: () => void;
}) {
  const [enviando, setEnviando] = useState(false);

  function alSubmit(e: FormEvent) {
    e.preventDefault();
    if (enviando) return;
    setEnviando(true);
    alConfirmar();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={alCerrar}>
      <div className="w-full max-w-lg rounded-md bg-white shadow-lg" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-stone-200 px-5 py-3">
          <h3 className="text-lg font-semibold text-cocina-oscuro">Reactivar sistema</h3>
          <button type="button" onClick={alCerrar} className="text-stone-500 hover:text-cocina-marron" aria-label="Cerrar">✕</button>
        </div>

        <form onSubmit={alSubmit} className="px-5 py-4 space-y-4">
          <div className="rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-900">
            <p className="font-semibold mb-1">Reactivación en cascada.</p>
            <p className="text-xs">
              El sistema vuelve a estar <strong>ACTIVO</strong> y disponible en los selectores al registrar accesos.{" "}
              {accesosVinculados === null
                ? "Sus accesos revocados también volverán a estar activos."
                : accesosVinculados === 0
                  ? <>No tiene accesos revocados — nada en cascada.</>
                  : <>Se reactivarán en cascada <strong>{accesosVinculados} acceso{accesosVinculados === 1 ? "" : "s"}</strong> vinculado{accesosVinculados === 1 ? "" : "s"}.</>}
            </p>
          </div>

          <div className="text-sm">
            <p className="text-stone-500 text-xs uppercase tracking-wider mb-1">Vas a reactivar</p>
            <p className="font-medium text-cocina-oscuro break-words">{sistema.nombre}</p>
            <p className="text-xs text-stone-500 break-words">{sistema.url_acceso}</p>
          </div>

          <div className="flex flex-col-reverse sm:flex-row justify-end gap-2 pt-2">
            <button type="button" onClick={alCerrar} className="boton-secundario" disabled={enviando}>
              Cancelar
            </button>
            <button
              type="submit"
              disabled={enviando}
              className="inline-flex items-center justify-center rounded-md bg-emerald-700 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-700 focus:ring-offset-2 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {enviando ? "Reactivando…" : "Reactivar sistema"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
