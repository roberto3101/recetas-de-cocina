import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  ListadoSistemas,
  ResultadoUsuariosAgregados,
  SistemaDisponible,
  UsuarioExterno,
} from "@/capacidades/accesos/tipos";

export default function PaginaAcceso() {
  const navegar = useNavigate();
  const [texto, setTexto] = useState("");
  const [textoBusqueda, setTextoBusqueda] = useState("");
  const [sistemaId, setSistemaId] = useState<string>("");
  const [hashesVisibles, setHashesVisibles] = useState<Record<string, boolean>>({});
  const [usuarioDetalles, setUsuarioDetalles] = useState<UsuarioExterno | null>(null);
  const [usuarioCambioPwd, setUsuarioCambioPwd] = useState<UsuarioExterno | null>(null);

  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);
  const [resultado, setResultado] = useState<ResultadoUsuariosAgregados | null>(null);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const r = await pedirJson<ListadoSistemas>("/cocina/sistemas");
        setSistemas(r.sistemas);
      } catch (e) {
        if (e instanceof Error) setError(e.message);
      }
    })();
  }, []);

  const cargar = useCallback(async () => {
    setCargando(true);
    setError(null);
    try {
      const params = new URLSearchParams({ tamano_pagina: "200" });
      if (textoBusqueda) params.set("q", textoBusqueda);
      const r = await pedirJson<ResultadoUsuariosAgregados>(`/cocina/usuarios-externos?${params.toString()}`);
      setResultado(r);
    } catch (e) {
      if (e instanceof Error) setError(e.message);
    } finally {
      setCargando(false);
    }
  }, [textoBusqueda]);

  useEffect(() => {
    void cargar();
  }, [cargar]);

  const usuariosFiltrados = useMemo<UsuarioExterno[]>(() => {
    if (!resultado) return [];
    if (!sistemaId) return resultado.usuarios;
    return resultado.usuarios.filter((u) => u.sistema_destino_id === sistemaId);
  }, [resultado, sistemaId]);

  function aplicarBusqueda() {
    setTextoBusqueda(texto.trim());
  }

  function urlBonita(u: UsuarioExterno) {
    return u.sistema_url_acceso || "—";
  }

  function llave(u: UsuarioExterno) {
    return `${u.sistema_destino_id}-${u.id_externo}`;
  }

  function alternarHash(u: UsuarioExterno) {
    const k = llave(u);
    setHashesVisibles((prev) => ({ ...prev, [k]: !prev[k] }));
  }

  async function copiarHash(u: UsuarioExterno) {
    if (!u.password_hash) return;
    try {
      await navigator.clipboard.writeText(u.password_hash);
    } catch {
      // ignore
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4 flex-wrap">
        <div>
          <h2 className="text-2xl font-semibold text-cocina-oscuro">Acceso</h2>
          <p className="text-sm text-stone-500">
            {resultado ? `${usuariosFiltrados.length} usuarios visibles · ${resultado.total_registros} totales` : "Cargando…"}
          </p>
        </div>
        <button onClick={() => navegar("/panel/registro")} className="boton-primario">+ Registrar acceso</button>
      </div>

      <form
        className="grid grid-cols-1 gap-3 sm:grid-cols-[2fr_1fr_auto]"
        onSubmit={(e) => { e.preventDefault(); aplicarBusqueda(); }}
      >
        <input
          type="text"
          placeholder="Buscar por correo electrónico…"
          className="campo-texto"
          value={texto}
          onChange={(e) => setTexto(e.target.value)}
        />
        <select
          className="campo-texto"
          value={sistemaId}
          onChange={(e) => setSistemaId(e.target.value)}
        >
          <option value="">Todas las URLs</option>
          {sistemas.map((s) => (
            <option key={s.id} value={s.id}>{s.nombre}</option>
          ))}
        </select>
        <button type="submit" className="boton-secundario">Buscar</button>
      </form>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">{error}</div>
      )}

      {resultado?.errores_por_sistema && resultado.errores_por_sistema.length > 0 && (
        <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-2 text-sm text-amber-900">
          <div className="font-medium mb-1">Algunos sistemas no respondieron:</div>
          <ul className="list-disc list-inside space-y-0.5">
            {resultado.errores_por_sistema.map((e) => (
              <li key={e.sistema_destino_id}><strong>{e.sistema_nombre}</strong>: {e.mensaje}</li>
            ))}
          </ul>
        </div>
      )}

      {usuarioDetalles && (
        <ModalDetalles usuario={usuarioDetalles} alCerrar={() => setUsuarioDetalles(null)} />
      )}
      {usuarioCambioPwd && (
        <ModalCambioPassword
          usuario={usuarioCambioPwd}
          alCerrar={() => setUsuarioCambioPwd(null)}
          alExito={() => { setUsuarioCambioPwd(null); void cargar(); }}
        />
      )}

      <div className="overflow-x-auto rounded-md border border-stone-200">
        <table className="min-w-full text-sm">
          <thead className="bg-stone-100 text-left text-xs uppercase tracking-wider text-stone-600">
            <tr>
              <th className="w-12 px-3 py-2">Orden</th>
              <th className="px-3 py-2">URL</th>
              <th className="px-3 py-2">Usuario</th>
              <th className="px-3 py-2">Clave (hash)</th>
              <th className="px-3 py-2 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody>
            {cargando && (
              <tr><td colSpan={5} className="px-3 py-6 text-center text-stone-500">Cargando…</td></tr>
            )}
            {!cargando && usuariosFiltrados.length === 0 && (
              <tr><td colSpan={5} className="px-3 py-6 text-center text-stone-500">No hay usuarios. Click en <strong>+ Registrar acceso</strong>.</td></tr>
            )}
            {!cargando && usuariosFiltrados.map((u, idx) => {
              const visible = !!hashesVisibles[llave(u)];
              const tieneHash = !!u.password_hash;
              return (
                <tr key={llave(u)} className="border-t border-stone-100 hover:bg-cocina-fondo align-top">
                  <td className="px-3 py-2 text-stone-500">{idx + 1}</td>
                  <td className="px-3 py-2 font-medium text-cocina-marron">
                    <a
                      href={`/cocina/sistemas/${u.sistema_destino_id}/usuarios/${encodeURIComponent(u.id_externo)}/autofill`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="hover:underline"
                      title={`Entrar a ${u.sistema_nombre} con clave guardada en el gestor`}
                    >
                      {urlBonita(u)} ↗
                    </a>
                    <div className="text-xs text-stone-400">{u.sistema_nombre}</div>
                  </td>
                  <td className="px-3 py-2 text-cocina-oscuro">
                    <div>{u.correo_electronico || "—"}</div>
                    {u.nombre_completo && <div className="text-xs text-stone-400">{u.nombre_completo}</div>}
                  </td>
                  <td className="px-3 py-2 font-mono text-xs text-stone-500 max-w-[280px]">
                    {tieneHash ? (
                      <div className="flex items-start gap-2">
                        <span className={`flex-1 ${visible ? "break-all" : "whitespace-nowrap"}`}>
                          {visible ? u.password_hash : "••••••••"}
                        </span>
                        <button
                          type="button"
                          onClick={() => alternarHash(u)}
                          className="text-stone-500 hover:text-cocina-marron shrink-0"
                          title={visible ? "Ocultar" : "Ver hash"}
                        >{visible ? "🙈" : "👁️"}</button>
                        <button
                          type="button"
                          onClick={() => copiarHash(u)}
                          className="text-stone-500 hover:text-cocina-marron shrink-0"
                          title="Copiar hash"
                        >📋</button>
                      </div>
                    ) : (
                      <span className="text-stone-300 whitespace-nowrap">— sin hash —</span>
                    )}
                  </td>
                  <td className="px-3 py-2 text-right whitespace-nowrap">
                    <div className="flex flex-wrap justify-end items-center gap-x-3 gap-y-1">
                      <button
                        type="button"
                        onClick={() => setUsuarioDetalles(u)}
                        className="text-xs text-cocina-marron hover:underline"
                        title="Ver detalles del usuario"
                      >Detalles</button>
                      <button
                        type="button"
                        onClick={() => setUsuarioCambioPwd(u)}
                        className="text-xs text-cocina-marron hover:underline"
                        title="Cambiar contraseña en el sistema externo"
                      >Cambiar clave</button>
                      {tieneHash && (
                        <a
                          href={`/cocina/sistemas/${u.sistema_destino_id}/usuarios/${encodeURIComponent(u.id_externo)}/autofill-hash`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-xs text-cocina-marron hover:underline"
                          title="Probar login enviando el hash como clave"
                        >Probar hash ↗</a>
                      )}
                      <span className="text-xs text-stone-400">{u.estado}</span>
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
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onClick={alCerrar}
    >
      <div
        className="w-full max-w-lg rounded-md bg-white shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-stone-200 px-5 py-3">
          <h3 className="text-lg font-semibold text-cocina-oscuro">{titulo}</h3>
          <button
            type="button"
            onClick={alCerrar}
            className="text-stone-500 hover:text-cocina-marron"
            aria-label="Cerrar"
          >✕</button>
        </div>
        <div className="px-5 py-4">{children}</div>
      </div>
    </div>
  );
}

function ModalDetalles({ usuario, alCerrar }: { usuario: UsuarioExterno; alCerrar: () => void }) {
  const filas: [string, string][] = [
    ["Sistema", usuario.sistema_nombre],
    ["URL", usuario.sistema_url_acceso],
    ["Correo", usuario.correo_electronico || "—"],
    ["Nombre", usuario.nombre_completo || "—"],
    ["Estado", usuario.estado || "—"],
    ["ID externo", usuario.id_externo],
    ["Hash en BD", usuario.password_hash || "— sin hash —"],
  ];
  return (
    <ModalContenedor titulo="Detalles del usuario" alCerrar={alCerrar}>
      <dl className="space-y-2 text-sm">
        {filas.map(([etiqueta, valor]) => (
          <div key={etiqueta} className="grid grid-cols-[110px_1fr] gap-3">
            <dt className="font-medium text-stone-500">{etiqueta}</dt>
            <dd className="break-all text-cocina-oscuro font-mono text-xs">{valor}</dd>
          </div>
        ))}
      </dl>
      <div className="mt-5 flex justify-end">
        <button type="button" onClick={alCerrar} className="boton-secundario">Cerrar</button>
      </div>
    </ModalContenedor>
  );
}

function ModalCambioPassword({ usuario, alCerrar, alExito }: { usuario: UsuarioExterno; alCerrar: () => void; alExito: () => void }) {
  const [passwordNueva, setPasswordNueva] = useState("");
  const [confirmar, setConfirmar] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const errorValidacion = useMemo(() => {
    if (passwordNueva.length < 8) return "Mínimo 8 caracteres";
    if (passwordNueva.length > 200) return "Máximo 200 caracteres";
    if (confirmar && passwordNueva !== confirmar) return "Las contraseñas no coinciden";
    return null;
  }, [passwordNueva, confirmar]);

  async function alGuardar(e: React.FormEvent) {
    e.preventDefault();
    if (errorValidacion || passwordNueva !== confirmar) {
      setError(errorValidacion ?? "Las contraseñas no coinciden");
      return;
    }
    setError(null);
    setEnviando(true);
    try {
      await enviarJson(
        `/cocina/sistemas/${usuario.sistema_destino_id}/usuarios/${encodeURIComponent(usuario.id_externo)}/password`,
        { password_nueva: passwordNueva },
        "PUT",
      );
      alExito();
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <ModalContenedor titulo="Cambiar contraseña" alCerrar={alCerrar}>
      <p className="text-sm text-stone-600 mb-3">
        <strong>{usuario.correo_electronico || usuario.id_externo}</strong> en {usuario.sistema_nombre}.
      </p>
      <div className="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-900 mb-4">
        ⚠️ Esto modifica la contraseña del usuario en la BD del sistema externo y guarda la nueva cifrada en el gestor para autofill.
      </div>
      <form onSubmit={alGuardar} className="space-y-3">
        <div>
          <label htmlFor="pwd-nueva" className="etiqueta-campo">Contraseña nueva</label>
          <input
            id="pwd-nueva"
            type="password"
            className="campo-texto"
            value={passwordNueva}
            onChange={(e) => setPasswordNueva(e.target.value)}
            autoComplete="new-password"
            required
            minLength={8}
            maxLength={200}
          />
        </div>
        <div>
          <label htmlFor="pwd-confirmar" className="etiqueta-campo">Confirmar</label>
          <input
            id="pwd-confirmar"
            type="password"
            className="campo-texto"
            value={confirmar}
            onChange={(e) => setConfirmar(e.target.value)}
            autoComplete="new-password"
            required
            minLength={8}
            maxLength={200}
          />
          {confirmar && errorValidacion && (
            <p className="text-xs text-red-600 mt-1">{errorValidacion}</p>
          )}
        </div>
        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
        )}
        <div className="flex justify-end gap-3 pt-2">
          <button type="button" className="boton-secundario" onClick={alCerrar} disabled={enviando}>Cancelar</button>
          <button type="submit" className="boton-primario" disabled={enviando || !!errorValidacion}>
            {enviando ? "Guardando…" : "Cambiar"}
          </button>
        </div>
      </form>
    </ModalContenedor>
  );
}
