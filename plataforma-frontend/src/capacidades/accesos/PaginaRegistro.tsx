import { FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";
import {
  ETIQUETAS_TIPO,
  ListadoSistemas,
  SistemaDisponible,
  TIPOS_ACCESO,
  TipoAcceso,
} from "@/capacidades/accesos/tipos";

const REGEX_CORREO = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type ErroresFormulario = Partial<{
  titulo: string;
  sistemaId: string;
  usuario: string;
  clave: string;
  puerto: string;
  observaciones: string;
}>;

export default function PaginaRegistro() {
  const navegar = useNavigate();
  const [sistemas, setSistemas] = useState<SistemaDisponible[]>([]);

  const [titulo, setTitulo] = useState("");
  const [tituloAutomatico, setTituloAutomatico] = useState(true);
  const [sistemaId, setSistemaId] = useState<string>("");
  const [usuario, setUsuario] = useState("");
  const [clave, setClave] = useState("");
  const [tipo, setTipo] = useState<TipoAcceso>("WEB");
  const [puerto, setPuerto] = useState<string>("");
  const [observaciones, setObservaciones] = useState("");

  const [tocado, setTocado] = useState<Record<string, boolean>>({});
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mensaje, setMensaje] = useState<string | null>(null);

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

  const sistemaActual = useMemo(
    () => sistemas.find((s) => s.id === sistemaId),
    [sistemas, sistemaId],
  );

  const errores = useMemo<ErroresFormulario>(() => {
    const e: ErroresFormulario = {};
    const tituloTrim = titulo.trim();
    if (!tituloTrim) e.titulo = "El título es obligatorio";
    else if (tituloTrim.length < 2) e.titulo = "Mínimo 2 caracteres";
    else if (tituloTrim.length > 200) e.titulo = "Máximo 200 caracteres";

    if (!sistemaId) e.sistemaId = "Selecciona una URL";

    const usuarioTrim = usuario.trim();
    if (!usuarioTrim) e.usuario = "El usuario es obligatorio";
    else if (usuarioTrim.length > 200) e.usuario = "Máximo 200 caracteres";
    else if (usuarioTrim.includes("@") && !REGEX_CORREO.test(usuarioTrim))
      e.usuario = "Formato de correo inválido";

    if (!clave) e.clave = "La clave es obligatoria";
    else if (clave.length > 500) e.clave = "Máximo 500 caracteres";

    const puertoTrim = puerto.trim();
    if (puertoTrim) {
      const n = Number(puertoTrim);
      if (!Number.isInteger(n) || n < 1 || n > 65535) e.puerto = "Entero entre 1 y 65535";
    }

    if (observaciones.length > 1000) e.observaciones = "Máximo 1000 caracteres";

    return e;
  }, [titulo, sistemaId, usuario, clave, puerto, observaciones]);

  const formularioValido = Object.keys(errores).length === 0;

  function marcarTocado(campo: string) {
    setTocado((t) => ({ ...t, [campo]: true }));
  }

  function mostrarError(campo: keyof ErroresFormulario): string | null {
    if (!tocado[campo]) return null;
    return errores[campo] ?? null;
  }

  async function alGuardar(evento: FormEvent) {
    evento.preventDefault();
    setError(null);
    setMensaje(null);
    setTocado({ titulo: true, sistemaId: true, usuario: true, clave: true, puerto: true, observaciones: true });
    if (!formularioValido || !sistemaActual) return;

    setEnviando(true);
    try {
      const cuerpo: Record<string, unknown> = {
        titulo: titulo.trim(),
        sistema_destino_id: sistemaActual.id,
        usuario_externo: usuario.trim(),
        password: clave,
        observaciones: observaciones.trim(),
        tipo,
      };
      if (puerto.trim()) cuerpo.puerto = Number(puerto);
      await enviarJson("/cocina/boveda/accesos", cuerpo);
      setMensaje(`Acceso guardado. Ya puedes entrar a ${sistemaActual.nombre} con click en la URL.`);
      setTimeout(() => navegar("/panel/acceso"), 800);
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  const claseCampo = (campo: keyof ErroresFormulario) =>
    `campo-texto ${mostrarError(campo) ? "border-red-400 focus:border-red-500" : ""}`;

  return (
    <div className="space-y-5">
      <h2 className="text-2xl font-semibold text-cocina-oscuro">Registrar acceso</h2>

      {sistemas.length === 0 && (
        <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          No tienes URLs registradas todavía. Ve a <a href="/panel/sistemas" className="underline">Sistemas</a> y crea una primero.
        </div>
      )}

      {mensaje && <div className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-2 text-sm text-emerald-800">{mensaje}</div>}
      {error && <div className="rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">{error}</div>}

      <form onSubmit={alGuardar} noValidate className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="lg:col-span-2">
          <label htmlFor="acceso-titulo" className="etiqueta-campo">Título</label>
          <input
            id="acceso-titulo"
            type="text"
            className={claseCampo("titulo")}
            value={titulo}
            onChange={(e) => { setTitulo(e.target.value); setTituloAutomatico(false); }}
            onBlur={() => marcarTocado("titulo")}
            maxLength={200}
            placeholder="ej. CRM Codeplex – gerencia"
          />
          {mostrarError("titulo") && <p className="text-xs text-red-600 mt-1">{mostrarError("titulo")}</p>}
        </div>

        <div className="lg:col-span-2">
          <label htmlFor="acceso-url" className="etiqueta-campo">URL del sistema</label>
          <select
            id="acceso-url"
            className={claseCampo("sistemaId")}
            value={sistemaId}
            onChange={(e) => {
              const nuevoId = e.target.value;
              setSistemaId(nuevoId);
              if (tituloAutomatico) {
                const s = sistemas.find((x) => x.id === nuevoId);
                setTitulo(s?.nombre ?? "");
              }
            }}
            onBlur={() => marcarTocado("sistemaId")}
          >
            <option value="">— Selecciona una URL —</option>
            {sistemas.map((s) => (
              <option key={s.id} value={s.id}>
                {s.nombre} — {s.url_acceso}
              </option>
            ))}
          </select>
          {mostrarError("sistemaId") && <p className="text-xs text-red-600 mt-1">{mostrarError("sistemaId")}</p>}
        </div>

        <div>
          <label htmlFor="acceso-usuario" className="etiqueta-campo">Usuario</label>
          <input
            id="acceso-usuario"
            type="text"
            className={claseCampo("usuario")}
            value={usuario}
            onChange={(e) => setUsuario(e.target.value)}
            onBlur={() => marcarTocado("usuario")}
            autoComplete="off"
            maxLength={200}
            placeholder="usuario o correo del sistema externo"
          />
          {mostrarError("usuario") && <p className="text-xs text-red-600 mt-1">{mostrarError("usuario")}</p>}
        </div>
        <div>
          <label htmlFor="acceso-clave" className="etiqueta-campo">Clave</label>
          <input
            id="acceso-clave"
            type="password"
            className={claseCampo("clave")}
            value={clave}
            onChange={(e) => setClave(e.target.value)}
            onBlur={() => marcarTocado("clave")}
            autoComplete="new-password"
            maxLength={500}
          />
          {mostrarError("clave") && <p className="text-xs text-red-600 mt-1">{mostrarError("clave")}</p>}
        </div>

        <div>
          <label htmlFor="acceso-puerto" className="etiqueta-campo">Puerto <span className="text-stone-400">(opcional)</span></label>
          <input
            id="acceso-puerto"
            type="number"
            min={1}
            max={65535}
            className={claseCampo("puerto")}
            value={puerto}
            onChange={(e) => setPuerto(e.target.value)}
            onBlur={() => marcarTocado("puerto")}
            placeholder="ej. 443, 22, 21"
          />
          {mostrarError("puerto") && <p className="text-xs text-red-600 mt-1">{mostrarError("puerto")}</p>}
        </div>
        <div>
          <label htmlFor="acceso-tipo" className="etiqueta-campo">Tipo</label>
          <select
            id="acceso-tipo"
            className="campo-texto"
            value={tipo}
            onChange={(e) => setTipo(e.target.value as TipoAcceso)}
          >
            {TIPOS_ACCESO.map((t) => (
              <option key={t} value={t}>{ETIQUETAS_TIPO[t]}</option>
            ))}
          </select>
        </div>

        <div className="lg:col-span-2">
          <label htmlFor="acceso-observaciones" className="etiqueta-campo">
            Observaciones <span className="text-stone-400">({observaciones.length}/1000)</span>
          </label>
          <textarea
            id="acceso-observaciones"
            rows={3}
            className={`${claseCampo("observaciones")} resize-y`}
            value={observaciones}
            onChange={(e) => setObservaciones(e.target.value)}
            onBlur={() => marcarTocado("observaciones")}
            maxLength={1000}
          />
          {mostrarError("observaciones") && <p className="text-xs text-red-600 mt-1">{mostrarError("observaciones")}</p>}
        </div>

        <div className="lg:col-span-2 flex items-center justify-end gap-3 pt-2">
          <button type="button" className="boton-secundario" onClick={() => navegar("/panel/acceso")}>Cancelar</button>
          <button type="submit" className="boton-primario" disabled={enviando || sistemas.length === 0}>
            {enviando ? "Guardando…" : "Guardar"}
          </button>
        </div>
      </form>
    </div>
  );
}
