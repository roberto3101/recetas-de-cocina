import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";

import { usarSesion } from "@/plataforma/identidad/usar_sesion";

export default function DisposicionCocina() {
  const navegar = useNavigate();
  const { perfil, cargando, cerrarSesion } = usarSesion();
  const [menuAbierto, setMenuAbierto] = useState(false);

  useEffect(() => {
    if (!cargando && !perfil) {
      navegar("/");
    }
  }, [cargando, perfil, navegar]);

  if (cargando) {
    return (
      <div className="flex h-screen items-center justify-center text-cocina-marron">
        Cargando…
      </div>
    );
  }
  if (!perfil) return null;

  async function alCerrar() {
    await cerrarSesion();
    navegar("/");
  }

  const claseEnlace = ({ isActive }: { isActive: boolean }) =>
    `block w-full text-left rounded-md px-4 py-2 text-sm font-medium transition-colors ${
      isActive
        ? "bg-cocina-marron text-white shadow"
        : "text-cocina-oscuro hover:bg-cocina-fondo"
    }`;

  const claseEnlaceMobile = ({ isActive }: { isActive: boolean }) =>
    `whitespace-nowrap rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
      isActive
        ? "bg-cocina-marron text-white shadow"
        : "text-cocina-oscuro hover:bg-cocina-fondo"
    }`;

  return (
    <div className="min-h-screen bg-stone-100">
      <header className="bg-cocina-marron text-white shadow">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-3 px-4 sm:px-6 py-3">
          <h1 className="text-base sm:text-lg font-semibold tracking-wide">Gestor</h1>
          <div className="flex items-center gap-2 sm:gap-4 text-sm">
            <span className="opacity-80 hidden sm:inline">{perfil.correo_electronico}</span>
            <button
              onClick={() => setMenuAbierto((v) => !v)}
              className="lg:hidden rounded border border-white/40 px-2 py-1"
              aria-label="Menú"
            >☰</button>
            <button
              onClick={alCerrar}
              className="rounded border border-white/40 px-2 sm:px-3 py-1 text-xs sm:text-sm hover:bg-white/10"
            >
              Salir
            </button>
          </div>
        </div>
        {menuAbierto && (
          <div className="lg:hidden bg-cocina-marron/95 px-4 pb-3 pt-1 text-xs opacity-90 border-t border-white/10">
            <span>{perfil.correo_electronico}</span>
          </div>
        )}
      </header>

      {/* Mobile: tabs scrollables; lg+: sidebar + main lado a lado */}
      <nav className="lg:hidden border-b border-stone-200 bg-white">
        <div className="mx-auto max-w-7xl overflow-x-auto px-4 py-2">
          <div className="flex gap-2 min-w-max">
            <NavLink to="/panel/acceso" className={claseEnlaceMobile} onClick={() => setMenuAbierto(false)}>Acceso</NavLink>
            <NavLink to="/panel/registro" className={claseEnlaceMobile} onClick={() => setMenuAbierto(false)}>Registrar</NavLink>
            <NavLink to="/panel/sistemas" className={claseEnlaceMobile} onClick={() => setMenuAbierto(false)}>Sistemas</NavLink>
            <NavLink to="/panel/cambiar-password" className={claseEnlaceMobile} onClick={() => setMenuAbierto(false)}>Cambiar clave</NavLink>
          </div>
        </div>
      </nav>

      <div className="mx-auto flex max-w-7xl gap-6 px-4 sm:px-6 py-4 sm:py-6">
        <aside className="hidden lg:block w-56 shrink-0">
          <nav className="space-y-1 rounded-md border border-stone-200 bg-white p-3 shadow-sm">
            <NavLink to="/panel/acceso" className={claseEnlace}>Acceso</NavLink>
            <NavLink to="/panel/registro" className={claseEnlace}>Registrar acceso</NavLink>
            <NavLink to="/panel/sistemas" className={claseEnlace}>Sistemas</NavLink>
            <div className="my-2 border-t border-stone-200" />
            <NavLink to="/panel/cambiar-password" className={claseEnlace}>Cambiar contraseña</NavLink>
          </nav>
        </aside>

        <main className="flex-1 rounded-md border border-stone-200 bg-white p-4 sm:p-6 shadow-sm min-h-[60vh]">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
