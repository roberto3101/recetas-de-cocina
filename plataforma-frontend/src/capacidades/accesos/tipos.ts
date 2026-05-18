export type SistemaDisponible = {
  id: string;
  codigo: string;
  nombre: string;
  url_acceso: string;
  url_login: string;
  nombre_campo_usuario: string;
  nombre_campo_password: string;
  metodo_login: string;
  estado: string;
};

export type ListadoSistemas = {
  sistemas: SistemaDisponible[];
  total: number;
};

export type TipoAcceso = "WEB" | "ESCRITORIO" | "FTP" | "OTRO";

export const TIPOS_ACCESO: TipoAcceso[] = ["WEB", "ESCRITORIO", "FTP", "OTRO"];

export const ETIQUETAS_TIPO: Record<TipoAcceso, string> = {
  WEB: "Web",
  ESCRITORIO: "Escritorio",
  FTP: "FTP",
  OTRO: "Otro",
};

export type AccesoGuardado = {
  id: string;
  titulo: string;
  sistema_destino_id: string;
  usuario_externo: string;
  observaciones: string;
  tipo: TipoAcceso;
  puerto?: number | null;
  estado: string;
  creado_en: string;
};

export type ListadoAccesos = {
  accesos: AccesoGuardado[];
  total: number;
};
