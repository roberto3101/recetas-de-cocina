export type RespuestaExitosa<T> = {
  exito: true;
  datos: T;
};

export type RespuestaError = {
  exito: false;
  error: string;
  codigo_error?: string;
};

export type RespuestaApi<T> = RespuestaExitosa<T> | RespuestaError;

export class ErrorApi extends Error {
  codigo: string;
  estadoHttp: number;
  constructor(mensaje: string, codigo: string, estadoHttp: number) {
    super(mensaje);
    this.codigo = codigo;
    this.estadoHttp = estadoHttp;
  }
}

async function procesarRespuesta<T>(respuesta: Response): Promise<T> {
  const tipo = respuesta.headers.get("Content-Type") ?? "";
  if (!tipo.includes("application/json")) {
    // 404 sigiloso del backend devuelve HTML; significa sesión perdida o ruta inexistente
    if (respuesta.status === 404) {
      throw new ErrorApi("Sesión expirada o ruta no encontrada. Vuelve a iniciar sesión.", "SESION_PERDIDA", 404);
    }
    throw new ErrorApi(`Respuesta inesperada del servidor (${respuesta.status})`, "RESPUESTA_NO_JSON", respuesta.status);
  }
  const cuerpo = (await respuesta.json()) as RespuestaApi<T>;
  if (!cuerpo.exito) {
    throw new ErrorApi(cuerpo.error, cuerpo.codigo_error ?? "DESCONOCIDO", respuesta.status);
  }
  return cuerpo.datos;
}

export async function pedirJson<T>(ruta: string): Promise<T> {
  const respuesta = await fetch(ruta, {
    method: "GET",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  return procesarRespuesta<T>(respuesta);
}

export async function enviarJson<T>(ruta: string, cuerpo: unknown, metodo: "POST" | "PUT" | "DELETE" = "POST"): Promise<T> {
  const respuesta = await fetch(ruta, {
    method: metodo,
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo),
  });
  return procesarRespuesta<T>(respuesta);
}

export async function enviarFormulario(ruta: string, datos: Record<string, string>): Promise<Response> {
  const cuerpo = new URLSearchParams();
  for (const [clave, valor] of Object.entries(datos)) {
    cuerpo.append(clave, valor);
  }
  return fetch(ruta, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: cuerpo.toString(),
  });
}
