import { InputHTMLAttributes, forwardRef, useState } from "react";

// CampoContrasena: input de contraseña con ojito para alternar visibilidad.
// Mantiene la firma estándar de <input> (value, onChange, autoComplete, etc.)
// para que se pueda reemplazar 1:1 cualquier <input type="password">.
//
// Props extra:
//   - validez: "ok" | "error" | undefined  → cambia el color del borde.
//   - claseCampoExtra: clases tailwind adicionales sobre .campo-texto.
//
// Por seguridad el botón ojito:
//   - tabIndex={-1} para que no quede atrapado en el flujo del Tab
//   - no submitea el form (type=button)
//   - aria-label dinámico para lectores de pantalla
//   - autocomplete sigue siendo lo que el caller indique (current-password,
//     new-password, off) — el toggle no lo afecta.
export type ValidezCampo = "ok" | "error" | undefined;

type Props = Omit<InputHTMLAttributes<HTMLInputElement>, "type"> & {
  validez?: ValidezCampo;
  claseCampoExtra?: string;
};

const CampoContrasena = forwardRef<HTMLInputElement, Props>(function CampoContrasena(
  { validez, claseCampoExtra = "", className, ...rest },
  ref,
) {
  const [visible, setVisible] = useState(false);

  const borde =
    validez === "error"
      ? "border-red-400 focus:border-red-500 focus:ring-red-300"
      : validez === "ok"
        ? "border-emerald-400 focus:border-emerald-500 focus:ring-emerald-300"
        : "";

  return (
    <div className="relative">
      <input
        {...rest}
        ref={ref}
        type={visible ? "text" : "password"}
        className={`campo-texto pr-10 ${borde} ${claseCampoExtra} ${className ?? ""}`.trim()}
      />
      <button
        type="button"
        tabIndex={-1}
        aria-label={visible ? "Ocultar contraseña" : "Mostrar contraseña"}
        onClick={() => setVisible((v) => !v)}
        className="absolute inset-y-0 right-0 flex items-center px-2 text-stone-400 hover:text-cocina-marron focus:outline-none focus:text-cocina-marron"
      >
        {visible ? <IconoOjoTachado /> : <IconoOjo />}
      </button>
    </div>
  );
});

export default CampoContrasena;

// Iconos inline para no agregar otra dependencia. Trazo simple para que
// combinen con el resto de la UI (no usamos librería de iconos en el repo).
function IconoOjo() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5">
      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function IconoOjoTachado() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5">
      <path d="M17.94 17.94A10.6 10.6 0 0 1 12 19c-6.5 0-10-7-10-7a17.7 17.7 0 0 1 3.18-4.19" />
      <path d="M9.9 4.24A10.8 10.8 0 0 1 12 4c6.5 0 10 7 10 7a17.5 17.5 0 0 1-2.16 3.19" />
      <path d="M1 1l22 22" />
      <path d="M14.12 14.12A3 3 0 1 1 9.88 9.88" />
    </svg>
  );
}
