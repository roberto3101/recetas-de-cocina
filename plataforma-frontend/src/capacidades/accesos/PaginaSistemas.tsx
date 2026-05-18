import { FormEvent, useCallback, useEffect, useState } from "react";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  ListadoSistemas,
  SistemaDisponible,
} from "@/capacidades/accesos/tipos";

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

  const [nombre, setNombre] = useState("");
  const [urlAcceso, setUrlAcceso] = useState("");
  const [enviando, setEnviando] = useState(false);

  const cargar = useCallback(async () => {
    setCargando(true);
    try {
      const r = await pedirJson<ListadoSistemas>("/cocina/sistemas");
      setSistemas(r.sistemas);
    } catch (e) {
      if (e instanceof Error) setError(e.message);
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => { void cargar(); }, [cargar]);

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

  async function eliminar(s: SistemaDisponible) {
    if (!confirm(`¿Eliminar "${s.nombre}"? Los accesos guardados que apunten a él dejarán de funcionar.`)) return;
    try {
      await enviarJson(`/cocina/sistemas/${s.id}`, undefined, "DELETE");
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
            Catálogo de URLs a las que vas a guardar credenciales. {sistemas.length} sistema(s).
          </p>
        </div>
        <button
          className="boton-primario"
          onClick={() => { setMostrandoForm(!mostrandoForm); setMensaje(null); setError(null); }}
        >
          {mostrandoForm ? "Cancelar" : "+ Nuevo sistema"}
        </button>
      </div>

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
            Sin sistemas. Toca <strong>+ Nuevo sistema</strong>.
          </div>
        )}
        {!cargando && sistemas.map((s) => (
          <div key={s.id} className="rounded-md border border-stone-200 bg-white p-3 space-y-1">
            <div className="font-medium text-cocina-oscuro">{s.nombre}</div>
            <div className="text-cocina-marron break-all text-xs">{s.url_acceso}</div>
            <div className="flex justify-end pt-1 border-t border-stone-100">
              <button onClick={() => eliminar(s)} className="text-xs text-red-700">Eliminar</button>
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
                Sin sistemas. Click en <strong>+ Nuevo sistema</strong>.
              </td></tr>
            )}
            {!cargando && sistemas.map((s) => (
              <tr key={s.id} className="border-t border-stone-100 hover:bg-cocina-fondo">
                <td className="px-3 py-2 font-medium text-cocina-oscuro">{s.nombre}</td>
                <td className="px-3 py-2 text-cocina-marron break-all text-xs">{s.url_acceso}</td>
                <td className="px-3 py-2 text-right whitespace-nowrap">
                  <button onClick={() => eliminar(s)} className="text-xs text-red-700 hover:underline">Eliminar</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
