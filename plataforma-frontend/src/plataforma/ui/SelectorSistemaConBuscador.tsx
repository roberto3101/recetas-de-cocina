import { useEffect, useMemo, useRef, useState } from "react";

import type { SistemaDisponible } from "@/capacidades/accesos/tipos";

// SelectorSistemaConBuscador: combobox para elegir un sistema entre muchos.
// Reemplaza el <select> nativo cuando hay >10 sistemas — escribir es más
// rápido que scrollear. Funciona con teclado completo:
//   - flechas arriba/abajo: navega resaltados
//   - Enter: confirma el resaltado
//   - Esc: cierra sin cambiar
//   - escribir: filtra por nombre o url
//
// Si ya hay un sistema seleccionado, se muestra como "pill" con una X para
// limpiarlo y volver a buscar. Si no hay nada, se ve como un input vacío.

type Props = {
  sistemas: SistemaDisponible[];
  valor: string;
  alCambiar: (id: string, sistema: SistemaDisponible | null) => void;
  alBlur?: () => void;
  conError?: boolean;
  placeholder?: string;
  idCampo?: string;
};

export default function SelectorSistemaConBuscador({
  sistemas,
  valor,
  alCambiar,
  alBlur,
  conError,
  placeholder = "Busca por nombre o URL…",
  idCampo,
}: Props) {
  const [abierto, setAbierto] = useState(false);
  const [consulta, setConsulta] = useState("");
  const [indiceResaltado, setIndiceResaltado] = useState(0);
  const refContenedor = useRef<HTMLDivElement>(null);
  const refInput = useRef<HTMLInputElement>(null);

  const seleccionado = useMemo(
    () => sistemas.find((s) => s.id === valor) ?? null,
    [sistemas, valor],
  );

  const filtrados = useMemo(() => {
    const q = consulta.trim().toLowerCase();
    if (!q) return sistemas;
    return sistemas.filter(
      (s) =>
        s.nombre.toLowerCase().includes(q) ||
        s.url_acceso.toLowerCase().includes(q) ||
        s.codigo.toLowerCase().includes(q),
    );
  }, [sistemas, consulta]);

  // Reset del índice resaltado cuando cambia el filtro.
  useEffect(() => {
    setIndiceResaltado(0);
  }, [consulta, abierto]);

  // Click fuera → cerrar.
  useEffect(() => {
    if (!abierto) return;
    function alClicGlobal(e: MouseEvent) {
      if (refContenedor.current && !refContenedor.current.contains(e.target as Node)) {
        setAbierto(false);
        alBlur?.();
      }
    }
    document.addEventListener("mousedown", alClicGlobal);
    return () => document.removeEventListener("mousedown", alClicGlobal);
  }, [abierto, alBlur]);

  function elegir(sis: SistemaDisponible) {
    alCambiar(sis.id, sis);
    setConsulta("");
    setAbierto(false);
  }

  function limpiar() {
    alCambiar("", null);
    setConsulta("");
    setAbierto(true);
    setTimeout(() => refInput.current?.focus(), 0);
  }

  function alTeclado(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setAbierto(true);
      setIndiceResaltado((i) => Math.min(i + 1, filtrados.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setIndiceResaltado((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      if (abierto && filtrados[indiceResaltado]) {
        e.preventDefault();
        elegir(filtrados[indiceResaltado]);
      }
    } else if (e.key === "Escape") {
      setAbierto(false);
    }
  }

  const claseBorde = conError
    ? "border-red-400 focus-within:border-red-500"
    : "border-gray-300 focus-within:border-cocina-marron";

  // Si ya hay seleccionado y NO está abierto → mostrar "pill" con X.
  if (seleccionado && !abierto) {
    return (
      <div ref={refContenedor} className="relative">
        <button
          type="button"
          id={idCampo}
          onClick={() => {
            setAbierto(true);
            setTimeout(() => refInput.current?.focus(), 0);
          }}
          className={`campo-texto flex w-full items-center justify-between text-left ${conError ? "border-red-400" : ""}`}
        >
          <span className="truncate">
            <span className="font-medium text-cocina-oscuro">{seleccionado.nombre}</span>
            <span className="ml-2 text-stone-400 text-xs">{seleccionado.url_acceso}</span>
          </span>
          <button
            type="button"
            tabIndex={-1}
            aria-label="Cambiar sistema"
            onClick={(e) => {
              e.stopPropagation();
              limpiar();
            }}
            className="ml-2 flex-shrink-0 text-stone-400 hover:text-cocina-marron"
          >
            <IconoX />
          </button>
        </button>
      </div>
    );
  }

  return (
    <div ref={refContenedor} className="relative">
      <div className={`flex items-center rounded-md border bg-white px-3 py-2 shadow-sm ${claseBorde}`}>
        <IconoLupa />
        <input
          ref={refInput}
          id={idCampo}
          type="text"
          autoComplete="off"
          className="ml-2 w-full bg-transparent text-sm outline-none placeholder:text-gray-400"
          placeholder={placeholder}
          value={consulta}
          onChange={(e) => {
            setConsulta(e.target.value);
            setAbierto(true);
          }}
          onFocus={() => setAbierto(true)}
          onKeyDown={alTeclado}
          role="combobox"
          aria-expanded={abierto}
          aria-controls="lista-sistemas-filtrada"
          aria-autocomplete="list"
        />
      </div>

      {abierto && (
        <ul
          id="lista-sistemas-filtrada"
          role="listbox"
          className="absolute z-20 mt-1 max-h-64 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg"
        >
          {filtrados.length === 0 ? (
            <li className="px-3 py-2 text-sm text-stone-500">
              Sin resultados para "{consulta}"
            </li>
          ) : (
            filtrados.map((s, i) => {
              const resaltado = i === indiceResaltado;
              const seleccionadoActual = s.id === valor;
              return (
                <li
                  key={s.id}
                  role="option"
                  aria-selected={seleccionadoActual}
                  onMouseDown={(e) => {
                    // mousedown en vez de click para ganarle al onBlur del input
                    e.preventDefault();
                    elegir(s);
                  }}
                  onMouseEnter={() => setIndiceResaltado(i)}
                  className={`cursor-pointer px-3 py-2 text-sm ${
                    resaltado ? "bg-cocina-fondo" : "bg-white"
                  } ${seleccionadoActual ? "font-medium text-cocina-marron" : "text-cocina-oscuro"}`}
                >
                  <div className="truncate">{s.nombre}</div>
                  <div className="truncate text-xs text-stone-500">{s.url_acceso}</div>
                </li>
              );
            })
          )}
        </ul>
      )}
    </div>
  );
}

function IconoLupa() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4 text-stone-400">
      <circle cx="11" cy="11" r="7" />
      <line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
  );
}

function IconoX() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}
