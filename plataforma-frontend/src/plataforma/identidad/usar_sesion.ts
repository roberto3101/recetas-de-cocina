import { useCallback, useEffect, useState } from "react";
import { ErrorApi, enviarJson, pedirJson } from "@/plataforma/red/cliente_api";

export type Perfil = {
  usuario_id: string;
  correo_electronico: string;
  segundo_factor_validado: boolean;
  expira_en: string;
};

export function usarSesion() {
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

  return { perfil, cargando, error, recargar, cerrarSesion };
}
