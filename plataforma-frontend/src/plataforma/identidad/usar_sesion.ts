import { createContext, useCallback, useContext, useEffect, useState, ReactNode, createElement } from "react";
import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";

export type Perfil = {
  usuario_id: string;
  correo_electronico: string;
  segundo_factor_validado: boolean;
  expira_en: string;
};

type EstadoSesion = {
  perfil: Perfil | null;
  cargando: boolean;
  error: string | null;
  recargar: () => Promise<void>;
  cerrarSesion: () => Promise<void>;
};

const ContextoSesion = createContext<EstadoSesion | null>(null);

// ProveedorSesion: hace UN solo fetch al perfil y lo comparte con todos los componentes.
// Reemplaza el patrón anterior donde cada componente llamaba usarSesion() y hacía su propio
// /cocina/identidad/perfil. Ahora hay un único fetch global por sesión.
export function ProveedorSesion({ children }: { children: ReactNode }) {
  const [perfil, setPerfil] = useState<Perfil | null>(null);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const recargar = useCallback(async () => {
    setCargando(true);
    setError(null);
    try {
      const datos = await pedirJson<Perfil>("/cocina/identidad/perfil");
      setPerfil(datos);
    } catch (e) {
      if (e instanceof ErrorApi && e.estadoHttp === 404) {
        setPerfil(null);
      } else if (e instanceof Error) {
        setError(e.message);
        setPerfil(null);
      }
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    void recargar();
  }, [recargar]);

  const cerrarSesion = useCallback(async () => {
    try {
      await enviarJson<{ sesion_cerrada: boolean }>("/cocina/identidad/cerrar-sesion", {});
    } finally {
      setPerfil(null);
    }
  }, []);

  const valor: EstadoSesion = { perfil, cargando, error, recargar, cerrarSesion };
  return createElement(ContextoSesion.Provider, { value: valor }, children);
}

// usarSesion: hook que consume el contexto. NO hace fetch propio.
// Si no hay Provider arriba en el árbol, lanza error (uso incorrecto).
export function usarSesion(): EstadoSesion {
  const ctx = useContext(ContextoSesion);
  if (!ctx) {
    throw new Error("usarSesion debe usarse dentro de <ProveedorSesion>");
  }
  return ctx;
}
