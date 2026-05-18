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

export type AccesoGuardado = {
  id: string;
  titulo: string;
  sistema_destino_id: string;
  usuario_externo: string;
  observaciones: string;
  estado: string;
  creado_en: string;
};

export type ListadoAccesos = {
  accesos: AccesoGuardado[];
  total: number;
};
