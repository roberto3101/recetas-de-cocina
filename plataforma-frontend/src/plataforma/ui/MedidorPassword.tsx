// MedidorPassword: muestra en tiempo real qué requisitos cumple una password.
// Refleja las reglas del backend (compartido/validaciones/password.go):
//   - mínimo 12 caracteres
//   - al menos una mayúscula
//   - al menos una minúscula
//   - al menos un dígito
//   - al menos un carácter especial (puntuación o símbolo)
//
// Adicionalmente, si se le pasa `confirmacion`, valida que coincida con la
// password y muestra un item extra "las contraseñas coinciden".
//
// El componente NO maneja state propio: el padre controla el valor.
// Exporta `evaluarPassword(valor)` para que el padre sepa si validar el
// submit y para alimentar la prop `validez` de <CampoContrasena>.

export type ResultadoEvaluacion = {
  longitud: boolean;
  mayuscula: boolean;
  minuscula: boolean;
  digito: boolean;
  especial: boolean;
  todoOk: boolean;
};

export function evaluarPassword(valor: string): ResultadoEvaluacion {
  const longitud = valor.length >= 12;
  const mayuscula = /[A-ZÁÉÍÓÚÑ]/.test(valor);
  const minuscula = /[a-záéíóúñ]/.test(valor);
  const digito = /\d/.test(valor);
  // \p{P} = puntuación Unicode, \p{S} = símbolos. Espejo de IsPunct||IsSymbol en Go.
  const especial = /[\p{P}\p{S}]/u.test(valor);
  const todoOk = longitud && mayuscula && minuscula && digito && especial;
  return { longitud, mayuscula, minuscula, digito, especial, todoOk };
}

type Props = {
  valor: string;
  confirmacion?: string;
  className?: string;
};

export default function MedidorPassword({ valor, confirmacion, className = "" }: Props) {
  const r = evaluarPassword(valor);
  // Solo mostramos el cheque de coincidencia si el padre realmente está
  // pidiendo confirmar (es decir, pasó el prop, aunque sea string vacío).
  const muestraConfirmacion = confirmacion !== undefined;
  const coincide = muestraConfirmacion && valor.length > 0 && valor === confirmacion;

  // Si la password está vacía, no mostramos nada — evita ruido al cargar.
  if (valor.length === 0 && (!muestraConfirmacion || confirmacion!.length === 0)) {
    return null;
  }

  return (
    <ul className={`mt-2 space-y-0.5 text-xs ${className}`} aria-live="polite">
      <Item ok={r.longitud} texto="Mínimo 12 caracteres" />
      <Item ok={r.mayuscula} texto="Una mayúscula" />
      <Item ok={r.minuscula} texto="Una minúscula" />
      <Item ok={r.digito} texto="Un dígito" />
      <Item ok={r.especial} texto="Un carácter especial (!@#…)" />
      {muestraConfirmacion && (
        <Item
          ok={coincide}
          texto={
            confirmacion!.length === 0
              ? "Confirma la contraseña"
              : coincide
                ? "Las contraseñas coinciden"
                : "Las contraseñas no coinciden"
          }
        />
      )}
    </ul>
  );
}

function Item({ ok, texto }: { ok: boolean; texto: string }) {
  return (
    <li className={`flex items-center gap-1.5 ${ok ? "text-emerald-700" : "text-stone-500"}`}>
      {ok ? <IconoCheck /> : <IconoPunto />}
      <span>{texto}</span>
    </li>
  );
}

function IconoCheck() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5 flex-shrink-0">
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function IconoPunto() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-3.5 w-3.5 flex-shrink-0">
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}
