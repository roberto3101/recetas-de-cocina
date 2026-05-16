import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

import { ErrorApi, enviarJson } from "@/plataforma/red/cliente_api";

export default function PaginaVerificarSegundoFactor() {
  const navegar = useNavigate();
  const [codigo, setCodigo] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function alValidar(evento: FormEvent) {
    evento.preventDefault();
    setError(null);
    setEnviando(true);
    try {
      await enviarJson<{ segundo_factor_validado: boolean }>(
        "/cocina/identidad/totp/validar",
        { codigo },
        "POST"
      );
      navegar("/panel/inventario");
    } catch (e) {
      if (e instanceof ErrorApi) setError(e.message);
      else if (e instanceof Error) setError(e.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div className="max-w-md">
      <h2 className="text-2xl font-semibold text-cocina-oscuro mb-1">Verificación de segundo factor</h2>
      <p className="text-sm text-stone-500 mb-5">
        Ingresa el código de 6 dígitos de tu aplicación de autenticación.
      </p>
      {error && (
        <div className="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      <form onSubmit={alValidar} className="space-y-4">
        <input
          type="text"
          inputMode="numeric"
          required
          maxLength={6}
          className="campo-texto font-mono tracking-widest text-center text-xl"
          value={codigo}
          onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ""))}
          autoFocus
        />
        <button type="submit" className="boton-primario w-full" disabled={enviando}>
          {enviando ? "Validando…" : "Validar"}
        </button>
      </form>
    </div>
  );
}
