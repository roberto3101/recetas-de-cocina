import { FormEvent, useCallback, useEffect, useState } from "react";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  ListadoSistemas,
  SistemaDisponible,
} from "@/capacidades/accesos/tipos";

export default function PaginaSistemas() {
  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mensaje, setMensaje] = useState<string | null>(null);
  const [mostrandoForm, setMostrandoForm] = useState(false);

  const [codigo, setCodigo] = useState("");
  const [nombre, setNombre] = useState("");
  const [urlAcceso, setUrlAcceso] = useState("");
  const [urlLogin, setUrlLogin] = useState("");
  const [campoUsuario, setCampoUsuario] = useState("correo_electronico");
  const [campoPassword, setCampoPassword] = useState("password");
  const [metodoLogin, setMetodoLogin] = useState("POST");
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
    setCodigo("");
    setNombre("");
    setUrlAcceso("");
    setUrlLogin("");
    setCampoUsuario("correo_electronico");
    setCampoPassword("password");
    setMetodoLogin("POST");
  }

  async function alGuardar(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setMensaje(null);
    setEnviando(true);
    try {
      await enviarJson("/cocina/sistemas", {
        codigo: codigo.trim().toLowerCase().replace(/\s+/g, "_"),
        nombre: nombre.trim(),
        url_acceso: urlAcceso.trim(),
        url_login: urlLogin.trim() || urlAcceso.trim(),
        nombre_campo_usuario: campoUsuario.trim(),
        nombre_campo_password: campoPassword.trim(),
        metodo_login: metodoLogin,
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
    if (!confirm(`¿Eliminar el sistema "${s.nombre}"? Los accesos guardados que apunten a él dejarán de funcionar.`)) return;
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
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
            <div>
              <label htmlFor="sis-nombre" className="etiqueta-campo">Nombre</label>
              <input id="sis-nombre" required type="text" className="campo-texto" value={nombre} onChange={(e) => setNombre(e.target.value)} placeholder="CRM Codeplex" />
            </div>
            <div>
              <label htmlFor="sis-codigo" className="etiqueta-campo">Código (slug)</label>
              <input id="sis-codigo" required type="text" className="campo-texto" value={codigo} onChange={(e) => setCodigo(e.target.value)} placeholder="crm_codeplex" />
            </div>
            <div className="lg:col-span-2">
              <label htmlFor="sis-url" className="etiqueta-campo">URL del sistema (donde se ve la página de login)</label>
              <input id="sis-url" required type="url" className="campo-texto" value={urlAcceso} onChange={(e) => setUrlAcceso(e.target.value)} placeholder="https://codeplex.pe/crm/admin/authentication" />
            </div>
            <div className="lg:col-span-2">
              <label htmlFor="sis-url-login" className="etiqueta-campo">URL del POST de login (opcional, si distinta a la anterior)</label>
              <input id="sis-url-login" type="url" className="campo-texto" value={urlLogin} onChange={(e) => setUrlLogin(e.target.value)} placeholder="(deja vacío para usar la URL del sistema)" />
            </div>
            <div>
              <label htmlFor="sis-campo-usuario" className="etiqueta-campo">Nombre del campo usuario en el form</label>
              <input id="sis-campo-usuario" type="text" className="campo-texto" value={campoUsuario} onChange={(e) => setCampoUsuario(e.target.value)} />
            </div>
            <div>
              <label htmlFor="sis-campo-password" className="etiqueta-campo">Nombre del campo clave en el form</label>
              <input id="sis-campo-password" type="text" className="campo-texto" value={campoPassword} onChange={(e) => setCampoPassword(e.target.value)} />
            </div>
            <div>
              <label htmlFor="sis-metodo" className="etiqueta-campo">Método</label>
              <select id="sis-metodo" className="campo-texto" value={metodoLogin} onChange={(e) => setMetodoLogin(e.target.value)}>
                <option value="POST">POST</option>
                <option value="GET">GET</option>
              </select>
            </div>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" className="boton-secundario" onClick={() => setMostrandoForm(false)} disabled={enviando}>Cancelar</button>
            <button type="submit" className="boton-primario" disabled={enviando}>
              {enviando ? "Guardando…" : "Guardar sistema"}
            </button>
          </div>
        </form>
      )}

      <div className="overflow-x-auto rounded-md border border-stone-200">
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
                <td className="px-3 py-2 font-medium text-cocina-oscuro">
                  {s.nombre}
                  <div className="text-xs text-stone-400">{s.codigo}</div>
                </td>
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
