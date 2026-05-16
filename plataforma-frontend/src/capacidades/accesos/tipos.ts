export type TipoAcceso = "WEB" | "ESCRITORIO" | "FTP" | "BASE_DATOS" | "API" | "OTRO";

export const TIPOS_ACCESO: TipoAcceso[] = ["WEB", "ESCRITORIO", "FTP", "BASE_DATOS", "API", "OTRO"];

export const ETIQUETAS_TIPO: Record<TipoAcceso, string> = {
  WEB: "Web",
  ESCRITORIO: "Escritorio",
  FTP: "FTP",
  BASE_DATOS: "Base de datos",
  API: "API",
  OTRO: "Otro",
};

export type AccesoListado = {
  id: string;
  titulo: string;
  tipo: TipoAcceso;
  url: string;
  puerto: number | null;
  observaciones: string;
  sistema_destino_id: string | null;
  permite_autoregistro: boolean;
  estado: string;
  creado_en: string;
};

export type ResultadoListadoAccesos = {
  entradas: AccesoListado[];
  total_registros: number;
  pagina: number;
  tamano_pagina: number;
};

export type SistemaDisponible = {
  id: string;
  codigo: string;
  nombre: string;
  url_acceso: string;
  soporta_autoregistro: boolean;
};

export type ListadoSistemas = {
  sistemas: SistemaDisponible[];
  total: number;
};

export type UsuarioExterno = {
  id_externo: string;
  correo_electronico: string;
  nombre_completo: string;
  estado: string;
  password_hash?: string;
  sistema_destino_id: string;
  sistema_codigo: string;
  sistema_nombre: string;
  sistema_url_acceso: string;
};

export type ErrorPorSistema = {
  sistema_destino_id: string;
  sistema_codigo: string;
  sistema_nombre: string;
  mensaje: string;
};

export type ResultadoUsuariosAgregados = {
  usuarios: UsuarioExterno[];
  total_registros: number;
  errores_por_sistema?: ErrorPorSistema[];
};
