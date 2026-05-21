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
  limite: number;
  offset: number;
};

// Opciones de orden expuestas al usuario. El value se manda como dos query
// params separados al backend (orden + direccion).
export type OpcionOrden = {
  etiqueta: string;
  campo: "creado_en" | "titulo";
  direccion: "asc" | "desc";
};

export const OPCIONES_ORDEN: OpcionOrden[] = [
  { etiqueta: "Más recientes", campo: "creado_en", direccion: "desc" },
  { etiqueta: "Más antiguos", campo: "creado_en", direccion: "asc" },
  { etiqueta: "Título A–Z", campo: "titulo", direccion: "asc" },
  { etiqueta: "Título Z–A", campo: "titulo", direccion: "desc" },
];

export const TAMANO_PAGINA_DEFAULT = 40;
