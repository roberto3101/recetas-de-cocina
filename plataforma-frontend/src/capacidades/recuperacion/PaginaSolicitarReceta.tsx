import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";

import { ErrorApi, enviarJson } from "@/plataforma/red/cliente_api";

// PaginaSolicitarReceta: formulario "olvidé mi contraseña" disfrazado.
// La UI habla de recetas; el backend habla de tokens. Solo se manda el
// correo electrónico (campo "ingrediente" en el JSON para matchear el
// vocabulario del cebo).
//
// El mensaje de respuesta es idéntico exista o no el correo — esto evita
// que un atacante use el formulario para enumerar correos válidos.
export default function PaginaSolicitarReceta() {
  const [ingrediente, setIngrediente] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [respuesta, setRespuesta] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function alEnviar(evento: FormEvent) {
    evento.preventDefault();
    setError(null);
    setRespuesta(null);
    setEnviando(true);
    try {
      const r = await enviarJson<{ mensaje: string }>("/buscar/receta-perdida", {
        ingrediente: ingrediente.trim(),
      });
      setRespuesta(r.mensaje);
    } catch (e) {
      if (e instanceof ErrorApi || e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main className="mx-auto max-w-md px-4 py-10">
      <h1 className="text-3xl font-bold text-cocina-marron">Receta perdida</h1>
      <p className="mt-2 text-cocina-oscuro text-sm">
        ¿No recuerdas tu receta secreta? Déjanos el ingrediente y te enviamos
        un enlace para reescribirla. El enlace dura 30 minutos.
      </p>

      {respuesta && (
        <div className="mt-5 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          {respuesta}
        </div>
      )}
      {error && (
        <div className="mt-5 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {error}
        </div>
      )}

      {!respuesta && (
        <form
          onSubmit={alEnviar}
          className="mt-6 rounded-md bg-cocina-fondo p-6 shadow-sm border border-amber-100 space-y-4"
        >
          <div>
            <label htmlFor="ingrediente" className="etiqueta-campo">
              Ingrediente principal
            </label>
            <input
              id="ingrediente"
              type="email"
              autoComplete="email"
              required
              className="campo-texto"
              value={ingrediente}
              onChange={(e) => setIngrediente(e.target.value)}
              placeholder="ej. ají@panca.pe"
            />
          </div>
          <button type="submit" className="boton-primario w-full" disabled={enviando}>
            {enviando ? "Buscando…" : "Recuperar mi receta"}
          </button>
        </form>
      )}

      <div className="mt-6 text-center">
        <Link to="/" className="text-sm text-stone-500 hover:text-cocina-marron underline">
          ← Volver al inicio
        </Link>
      </div>
    </main>
  );
}
