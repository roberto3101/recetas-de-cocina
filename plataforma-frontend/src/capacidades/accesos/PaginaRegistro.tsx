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
  const [puerto, setPuerto] = useState("");
  const [tipo, setTipo] = useState<TipoAcceso>("WEB");
  const [observaciones, setObservaciones] = useState("");
  const [crearSiNoExiste, setCrearSiNoExiste] = useState(false);

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
    else if (tituloTrim.length < 3) e.titulo = "Mínimo 3 caracteres";
    else if (tituloTrim.length > 200) e.titulo = "Máximo 200 caracteres";

    if (!sistemaId) e.sistemaId = "Selecciona una URL";

    const usuarioTrim = usuario.trim();
    if (!usuarioTrim) e.usuario = "El usuario es obligatorio";
    else if (usuarioTrim.length < 3) e.usuario = "Mínimo 3 caracteres";
    else if (usuarioTrim.length > 200) e.usuario = "Máximo 200 caracteres";
    else if (!REGEX_CORREO.test(usuarioTrim)) e.usuario = "Formato de correo inválido";

    if (!clave) e.clave = "La clave es obligatoria";
    else if (clave.length < 4) e.clave = "Mínimo 4 caracteres";
    else if (clave.length > 200) e.clave = "Máximo 200 caracteres";

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
    if (!formularioValido) return;
    if (!sistemaActual) return;

    setEnviando(true);
    try {
      if (crearSiNoExiste && sistemaActual.soporta_autoregistro) {
        await enviarJson(`/cocina/sistemas/${sistemaActual.id}/usuarios`, {
          correo_electronico: usuario.trim(),
          password_plana: clave,
          nombre_completo: titulo.trim(),
        }, "POST");
        setMensaje(`Usuario creado en ${sistemaActual.nombre} y credencial guardada para autofill.`);
      } else {
        await enviarJson(`/cocina/sistemas/${sistemaActual.id}/credencial-externa`, {
          correo: usuario.trim(),
          password: clave,
        }, "POST");
        setMensaje(`Credencial guardada. Ya puedes entrar a ${sistemaActual.nombre} con click en la URL.`);
      }
      setTimeout(() => navegar("/panel/acceso"), 900);
    } catch (e) {
      if (e instanceof ErrorApi) {
        if (e.codigo === "USUARIO_EXTERNO_NO_ENCONTRADO" && sistemaActual.soporta_autoregistro) {
          setError(`Ese correo no existe en ${sistemaActual.nombre}. Marca "Crear nuevo" si quieres registrarlo.`);
        } else {
          setError(e.message);
        }
      } else if (e instanceof Error) {
        setError(e.message);
      }
    } finally {
      setEnviando(false);
    }
  }

  const claseCampo = (campo: keyof ErroresFormulario) =>
    `campo-texto ${mostrarError(campo) ? "border-red-400 focus:border-red-500" : ""}`;

  return (
    <div className="space-y-5">
      <h2 className="text-2xl font-semibold text-cocina-oscuro">Registrar acceso</h2>

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
          />
          {mostrarError("titulo") && <p className="text-xs text-red-600 mt-1">{mostrarError("titulo")}</p>}
        </div>

        <div className="lg:col-span-2">
          <label htmlFor="acceso-url" className="etiqueta-campo">URL</label>
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
                {s.url_acceso} ({s.nombre})
              </option>
            ))}
          </select>
          {mostrarError("sistemaId") && <p className="text-xs text-red-600 mt-1">{mostrarError("sistemaId")}</p>}
        </div>

        <div>
          <label htmlFor="acceso-usuario" className="etiqueta-campo">Usuario</label>
          <input
            id="acceso-usuario"
            type="email"
            className={claseCampo("usuario")}
            value={usuario}
            onChange={(e) => setUsuario(e.target.value)}
            onBlur={() => marcarTocado("usuario")}
            autoComplete="off"
            maxLength={200}
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
            maxLength={200}
          />
          {mostrarError("clave") && <p className="text-xs text-red-600 mt-1">{mostrarError("clave")}</p>}
        </div>

        <div>
          <label htmlFor="acceso-puerto" className="etiqueta-campo">Puerto</label>
          <input
            id="acceso-puerto"
            type="number"
            min={1}
            max={65535}
            className={claseCampo("puerto")}
            value={puerto}
            onChange={(e) => setPuerto(e.target.value)}
            onBlur={() => marcarTocado("puerto")}
            placeholder="opcional"
          />
          {mostrarError("puerto") && <p className="text-xs text-red-600 mt-1">{mostrarError("puerto")}</p>}
        </div>
        <div>
          <label htmlFor="acceso-tipo" className="etiqueta-campo">Tipo</label>
          <select id="acceso-tipo" className="campo-texto" value={tipo} onChange={(e) => setTipo(e.target.value as TipoAcceso)}>
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

        {sistemaActual?.soporta_autoregistro && (
          <div className="lg:col-span-2 flex items-start gap-2 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm">
            <input
              id="acceso-crear-nuevo"
              type="checkbox"
              checked={crearSiNoExiste}
              onChange={(e) => setCrearSiNoExiste(e.target.checked)}
              className="h-4 w-4 mt-0.5"
            />
            <label htmlFor="acceso-crear-nuevo" className="text-cocina-oscuro">
              Crear este usuario en la BD de <strong>{sistemaActual.nombre}</strong> (sólo si no existe ya).
              <div className="text-xs text-stone-500">Si lo desactivas, solo se guardará la contraseña asociada al usuario que ya existe.</div>
            </label>
          </div>
        )}

        <div className="lg:col-span-2 flex items-center justify-end gap-3 pt-2">
          <button type="button" className="boton-secundario" onClick={() => navegar("/panel/acceso")}>Cancelar</button>
          <button type="submit" className="boton-primario" disabled={enviando}>
            {enviando ? "Guardando…" : "Guardar"}
          </button>
        </div>
      </form>
    </div>
  );
}
