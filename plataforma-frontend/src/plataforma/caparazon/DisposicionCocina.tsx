import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useEffect } from "react";

import { usarSesion } from "@/plataforma/identidad/usar_sesion";

export default function DisposicionCocina() {
  const navegar = useNavigate();
  const { perfil, cargando, cerrarSesion } = usarSesion();

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

  return (
    <div className="min-h-screen bg-stone-100">
      <header className="bg-cocina-marron text-white shadow">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
          <h1 className="text-lg font-semibold tracking-wide">Gestor de accesos</h1>
          <div className="flex items-center gap-4 text-sm">
            <span className="opacity-80">{perfil.correo_electronico}</span>
            <button onClick={alCerrar} className="rounded border border-white/40 px-3 py-1 hover:bg-white/10">
              Cerrar sesión
            </button>
          </div>
        </div>
      </header>

      <div className="mx-auto flex max-w-7xl gap-6 px-6 py-6">
        <aside className="w-56 shrink-0">
          <nav className="space-y-1 rounded-md border border-stone-200 bg-white p-3 shadow-sm">
            <NavLink to="/panel/acceso" className={claseEnlace}>Acceso</NavLink>
            <NavLink to="/panel/registro" className={claseEnlace}>Registrar acceso</NavLink>
            <NavLink to="/panel/sistemas" className={claseEnlace}>Sistemas</NavLink>
            <div className="my-2 border-t border-stone-200" />
            <NavLink to="/panel/cambiar-password" className={claseEnlace}>Cambiar contraseña</NavLink>
          </nav>
        </aside>

        <main className="flex-1 rounded-md border border-stone-200 bg-white p-6 shadow-sm min-h-[60vh]">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
