// Genera documentos/guia-tecnica.docx — explica BD, cifrado, seguridad
const fs = require("fs");
const path = require("path");
const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  HeadingLevel, AlignmentType, BorderStyle, WidthType, ShadingType,
  PageOrientation, LevelFormat, PageBreak, Header, Footer, PageNumber,
} = require("docx");

const border = { style: BorderStyle.SINGLE, size: 6, color: "BBBBBB" };
const borders = { top: border, bottom: border, left: border, right: border };

const COLOR_PRIMARIO = "8B3A2A"; // marrón cocina
const COLOR_SECUNDARIO = "5C2E20";
const FILL_ZEBRA = "FAF5F0";
const FILL_HEADER = "F0DECC";

function H1(t) { return new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun({ text: t, color: COLOR_PRIMARIO })] }); }
function H2(t) { return new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun({ text: t, color: COLOR_SECUNDARIO })] }); }
function H3(t) { return new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun({ text: t })] }); }
function P(...runs) { return new Paragraph({ spacing: { after: 120 }, children: runs.map((r) => typeof r === "string" ? new TextRun(r) : r) }); }
function Bold(t) { return new TextRun({ text: t, bold: true }); }
function Code(t) { return new TextRun({ text: t, font: "Consolas", color: COLOR_SECUNDARIO }); }
function Bullet(text) {
  return new Paragraph({
    numbering: { reference: "bullets", level: 0 },
    children: [new TextRun(text)],
  });
}

function cell(text, opts = {}) {
  return new TableCell({
    borders,
    width: { size: opts.width || 1872, type: WidthType.DXA },
    shading: opts.shading ? { fill: opts.shading, type: ShadingType.CLEAR } : undefined,
    margins: { top: 80, bottom: 80, left: 120, right: 120 },
    children: (Array.isArray(text) ? text : [text]).map((t) =>
      typeof t === "string"
        ? new Paragraph({ children: [new TextRun({ text: t, bold: !!opts.bold })] })
        : t
    ),
  });
}

function tabla(columnas, filasDatos, anchosCol) {
  const anchos = anchosCol || columnas.map(() => Math.floor(9360 / columnas.length));
  const tabla = new Table({
    width: { size: 9360, type: WidthType.DXA },
    columnWidths: anchos,
    rows: [
      new TableRow({
        children: columnas.map((c, i) => cell(c, { width: anchos[i], shading: FILL_HEADER, bold: true })),
      }),
      ...filasDatos.map((fila, idx) =>
        new TableRow({
          children: fila.map((v, i) => cell(v, { width: anchos[i], shading: idx % 2 === 0 ? FILL_ZEBRA : undefined })),
        })
      ),
    ],
  });
  return tabla;
}

function diagramaMonospace(lineas) {
  // Bloque ASCII art con borde — usamos una tabla 1x1 con monospace
  const contenido = lineas.map((l) =>
    new Paragraph({
      children: [new TextRun({ text: l, font: "Consolas", size: 18 })],
      spacing: { before: 0, after: 0 },
    })
  );
  return new Table({
    width: { size: 9360, type: WidthType.DXA },
    columnWidths: [9360],
    rows: [
      new TableRow({
        children: [
          new TableCell({
            borders,
            width: { size: 9360, type: WidthType.DXA },
            shading: { fill: "F5F5F0", type: ShadingType.CLEAR },
            margins: { top: 200, bottom: 200, left: 240, right: 240 },
            children: contenido,
          }),
        ],
      }),
    ],
  });
}

const doc = new Document({
  creator: "roberto3101",
  title: "Guía técnica — Gestor de accesos Codeplex",
  description: "Cómo se guardan los datos, las tablas, dónde viven y consideraciones de seguridad",
  styles: {
    default: { document: { run: { font: "Calibri", size: 22 } } },
    paragraphStyles: [
      { id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 36, bold: true, font: "Calibri", color: COLOR_PRIMARIO },
        paragraph: { spacing: { before: 300, after: 200 }, outlineLevel: 0 } },
      { id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 28, bold: true, font: "Calibri", color: COLOR_SECUNDARIO },
        paragraph: { spacing: { before: 220, after: 140 }, outlineLevel: 1 } },
      { id: "Heading3", name: "Heading 3", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 24, bold: true, font: "Calibri" },
        paragraph: { spacing: { before: 160, after: 100 }, outlineLevel: 2 } },
    ],
  },
  numbering: {
    config: [
      { reference: "bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
    ],
  },
  sections: [{
    properties: {
      page: {
        size: { width: 12240, height: 15840 },
        margin: { top: 1440, right: 1440, bottom: 1440, left: 1440 },
      },
    },
    headers: {
      default: new Header({
        children: [new Paragraph({
          alignment: AlignmentType.RIGHT,
          children: [new TextRun({ text: "Gestor de accesos — Guía técnica", color: "888888", size: 18 })],
        })],
      }),
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          alignment: AlignmentType.CENTER,
          children: [
            new TextRun({ text: "Página ", size: 18, color: "888888" }),
            new TextRun({ children: [PageNumber.CURRENT], size: 18, color: "888888" }),
          ],
        })],
      }),
    },
    children: [
      // PORTADA
      new Paragraph({
        spacing: { before: 2000, after: 200 },
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "Gestor de accesos", bold: true, size: 56, color: COLOR_PRIMARIO })],
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { after: 600 },
        children: [new TextRun({ text: "Guía técnica · cómo se guarda y se protege la información", size: 28, color: "555555" })],
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "Sistemas Unificados Codeplex · 2026", italics: true, color: "888888" })],
      }),
      new Paragraph({ children: [new PageBreak()] }),

      // 1. PARA QUÉ SIRVE
      H1("1. ¿Para qué sirve este sistema?"),
      P("El gestor de accesos es una ", Bold("bóveda personal de contraseñas"), " para los sistemas internos de Codeplex. Funciona como 1Password o Bitwarden, pero corre en infraestructura propia y bajo control del operador."),
      P("Resumen en una línea: ", Bold("guarda credenciales cifradas y permite copiarlas o abrir el sistema con 1 click sin tipear la clave.")),

      H2("Qué hace"),
      Bullet("Registrar URLs de sistemas a los que el operador necesita acceder (CRM, panel de control, FTP, etc.)."),
      Bullet("Guardar usuario + contraseña + observaciones para cada URL, todo cifrado en base de datos."),
      Bullet("Mostrar la lista de accesos, copiar la clave al portapapeles con un click, abrir el sistema en pestaña nueva."),
      Bullet("Activar / desactivar accesos sin perder el historial (eliminación lógica)."),

      H2("Qué NO hace"),
      Bullet("No se conecta a las bases de datos de los sistemas externos. No lee usuarios remotos."),
      Bullet("No envía emails de recuperación (es uso interno, 1 operador)."),
      Bullet("No hace login automático cross-origin (limitación de los navegadores con sistemas que usan CSRF tipo Laravel)."),

      // 2. ARQUITECTURA
      H1("2. Arquitectura general"),
      P("El sistema tiene tres piezas que viven en el servidor de producción:"),
      diagramaMonospace([
        "  ┌──────────────────────────────────────────────────────────────┐",
        "  │  Usuario (browser / PWA en móvil)                            │",
        "  └────────────────────────────┬─────────────────────────────────┘",
        "                               │ HTTPS (Let's Encrypt)",
        "                               ▼",
        "  ┌──────────────────────────────────────────────────────────────┐",
        "  │  nginx 1.24                                                  │",
        "  │  · Sirve el SPA estático (React + Vite + PWA)                │",
        "  │  · Reverse-proxy a /cocina, /buscar, /salud al backend       │",
        "  │  · Headers de seguridad (HSTS, CSP, X-Frame-Options, etc.)   │",
        "  └────────────────────────────┬─────────────────────────────────┘",
        "                               │ HTTP local (127.0.0.1:8081)",
        "                               ▼",
        "  ┌──────────────────────────────────────────────────────────────┐",
        "  │  Backend Go (pm2: gestor-codeplex-api)                       │",
        "  │  · Capacidades: identidad, catalogo_sistemas, boveda         │",
        "  │  · Cifrado AES-256-GCM con KEK del .env                      │",
        "  │  · Sesiones, auditoría, rate limit, validaciones             │",
        "  └────────────────────────────┬─────────────────────────────────┘",
        "                               │ pgx (driver Postgres compatible)",
        "                               ▼",
        "  ┌──────────────────────────────────────────────────────────────┐",
        "  │  CockroachDB 127.0.0.1:26257 (base sistemas_unificados)     │",
        "  │  · 6 tablas: usuario, sesion_global, sistema_destino,        │",
        "  │    acceso_guardado, auditoria_accion, usuario_totp           │",
        "  └──────────────────────────────────────────────────────────────┘",
      ]),
      P("Todo corre en ", Code("46.225.174.213"), " (servidor Ubuntu del operador). El dominio público es ", Code("recetas-cocina.duckdns.org"), ", con un SSL automático que Let's Encrypt renueva cada 90 días."),

      // 3. CÓMO SE GUARDAN LOS DATOS
      H1("3. Cómo se guardan los datos"),
      P("Hay 6 tablas en la base de datos del gestor. Cada una guarda algo distinto y con distinto nivel de protección. Resumen rápido:"),
      tabla(
        ["Tabla", "Qué guarda", "Cómo se protege"],
        [
          ["usuario", "El operador del gestor (correo y hash de su contraseña).", "Argon2id (irreversible). Ni yo ni el operador pueden leer la clave en plano una vez guardada."],
          ["sesion_global", "Cookie de sesión activa cuando estás logueado.", "Solo se guarda el hash SHA-256 del token. Si te roban la BD no pueden suplantar sesiones."],
          ["usuario_totp", "Secreto TOTP del operador (si activa 2FA).", "AES-256-GCM cifrado con el KEK."],
          ["sistema_destino", "Catálogo de URLs a las que vas a guardar credenciales.", "Texto plano. Es metadata pública (URLs, nombre del sistema)."],
          ["acceso_guardado", "Las credenciales de la bóveda: usuario y clave de cada acceso.", "Usuario, clave y observaciones cifrados AES-256-GCM. Hash HMAC-SHA256 para índice único."],
          ["auditoria_accion", "Bitácora de cada acción que hace el operador (login, crear acceso, etc.).", "Texto plano. No registra valores sensibles, solo eventos."],
        ],
        [1872, 3600, 3888]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("3.1 Tabla acceso_guardado en detalle"),
      P("Esta es la tabla que tiene la información sensible. Cada fila es un acceso guardado por el operador:"),
      tabla(
        ["Columna", "Tipo", "¿Cifrado?", "Qué contiene"],
        [
          ["id", "UUID", "—", "Identificador único de la fila"],
          ["titulo", "TEXT", "No", "Nombre que le pones al acceso (ej. 'CRM Codeplex – gerencia')"],
          ["sistema_destino_id", "UUID", "—", "Referencia a la URL del catálogo"],
          ["usuario_externo", "BYTES", Bold("Sí · AES-256-GCM"), "Correo o usuario que va al form de login externo"],
          ["usuario_externo_hash", "BYTES", "—", "HMAC-SHA256 determinístico solo para índice único (no revela el valor)"],
          ["password_cifrada", "BYTES", Bold("Sí · AES-256-GCM"), "La contraseña que se guarda para autofill"],
          ["observaciones", "BYTES", Bold("Sí · AES-256-GCM"), "Notas libres del operador (PINs, preguntas de seguridad, etc.)"],
          ["tipo", "TEXT", "No", "WEB / ESCRITORIO / FTP / OTRO"],
          ["puerto", "INT2", "No", "Puerto si aplica (FTP=21, SSH=22, etc.). Opcional"],
          ["estado", "TEXT", "—", "ACTIVO / REVOCADO / ELIMINADO"],
          ["creado_en, creado_por, …", "—", "—", "Auditoría: cuándo y quién"],
        ],
        [1750, 1100, 1700, 4810]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("3.2 ¿Por qué hay un hash determinístico además del cifrado?"),
      P("Hay un detalle técnico interesante: AES-GCM ", Bold("no es determinístico"), ". Cada vez que se cifra el mismo valor produce bytes diferentes (porque usa un nonce aleatorio). Eso es bueno para seguridad — un atacante no puede saber que dos accesos comparten el mismo usuario solo viendo los bytes."),
      P("Pero entonces, ¿cómo evitamos que el operador guarde el mismo usuario dos veces para el mismo sistema? No podríamos comparar el cifrado."),
      P("Solución: además del cifrado, se guarda un ", Bold("HMAC-SHA256"), " del usuario (que sí es determinístico). El índice único de la tabla usa ese hash, no el valor cifrado. El hash no revela el valor pero permite comparar."),

      H2("3.3 El KEK (clave que cifra todo)"),
      P("La llave maestra que descifra el contenido de la bóveda se llama ", Bold("KEK"), " (Key Encryption Key). Es una clave AES-256 de 32 bytes."),
      Bullet("Vive en el archivo /root/gestor-codeplex/.env del servidor."),
      Bullet("Permisos chmod 600 — solo el usuario root del servidor puede leerla."),
      Bullet("NO se guarda en la base de datos. Si alguien dumpea la BD pero no tiene el KEK, no puede descifrar nada."),
      Bullet("Si se pierde el KEK, todas las claves cifradas se vuelven irrecuperables. Es responsabilidad del operador respaldarlo (password manager personal, por ejemplo)."),

      new Paragraph({ children: [new PageBreak()] }),

      // 4. FLUJO DE CIFRADO
      H1("4. Flujo de cifrado paso a paso"),
      H2("4.1 Al guardar un acceso"),
      diagramaMonospace([
        "  1. Operador escribe en el form:           usuario='juan@x.com'",
        "                                            clave='SuperSecreta!'",
        "                                            observaciones='cuenta gerencia'",
        "                                                  │",
        "                                                  ▼",
        "  2. Backend recibe el JSON                 (ya viajó por HTTPS)",
        "                                                  │",
        "                                                  ▼",
        "  3. Lee el KEK del .env                    KEK = 32 bytes random",
        "                                                  │",
        "          ┌───────────────────────────────────────┴───────────────────────────┐",
        "          ▼                                                                   ▼",
        "  4a. Cifra cada campo con AES-256-GCM:                          4b. Hashea usuario:",
        "      usuario_cifrado   = AES(KEK, 'juan@x.com')                     hash = HMAC-SHA256(KEK, 'juan@x.com')",
        "      password_cifrada  = AES(KEK, 'SuperSecreta!')                  (determinístico → sirve para UNIQUE INDEX)",
        "      observaciones_cif = AES(KEK, 'cuenta gerencia')",
        "                                                  │",
        "                                                  ▼",
        "  5. INSERT INTO acceso_guardado (...)        ← solo bytes opacos en BD",
      ]),

      H2("4.2 Al leer un acceso"),
      diagramaMonospace([
        "  1. Frontend pide GET /cocina/boveda/accesos          (cookie de sesión)",
        "                       │",
        "                       ▼",
        "  2. Backend valida sesión + 2FA                       (404 sigiloso si no)",
        "                       │",
        "                       ▼",
        "  3. Lee filas de acceso_guardado                      (bytes cifrados)",
        "                       │",
        "                       ▼",
        "  4. Por cada fila, descifra al vuelo con el KEK",
        "      usuario     = decrypt_AES(KEK, usuario_cifrado)",
        "      observaciones = decrypt_AES(KEK, observaciones_cif)",
        "      (la clave NO se descifra en /accesos; solo en /bookmarklet por petición explícita)",
        "                       │",
        "                       ▼",
        "  5. Devuelve JSON con valores en plano                (HTTPS al frontend)",
        "                       │",
        "                       ▼",
        "  6. Frontend muestra la lista                         (clave queda como ••••••••)",
      ]),

      H2("4.3 Tolerancia a fallos de descifrado"),
      P("Si una fila tiene bytes que no se pueden descifrar (KEK rotada, dato corrupto, residuos de pruebas), el listado ", Bold("no falla entero"), ". Esa fila se omite con un warning en el log del backend; las demás se muestran normalmente. Esto evita que un solo dato malo bloquee toda la pantalla."),

      new Paragraph({ children: [new PageBreak()] }),

      // 5. SEGURIDAD
      H1("5. Consideraciones de seguridad"),
      P("El sistema usa varias capas de defensa. Si una falla, las otras siguen protegiendo. Lista resumida:"),

      H2("5.1 Autenticación"),
      tabla(
        ["Defensa", "Cómo funciona"],
        [
          ["Argon2id en password del operador", "La clave del operador NO se guarda nunca; solo se guarda un hash Argon2id (irreversible y resistente a GPUs/ASICs). Ni yo ni un atacante con la BD pueden recuperar la clave en plano."],
          ["Login disfrazado de blog", "La página pública parece un blog de recetas. Un atacante casual no sabe que ahí dentro hay un login. Cualquier ruta /cocina/* sin sesión devuelve 404 idéntico al de rutas inexistentes."],
          ["Anti-timing attack", "Si el correo NO existe, igual se ejecuta Argon2id contra un hash dummy. El tiempo de respuesta entre 'usuario existe' y 'usuario no existe' es indistinguible — bloquea enumeración."],
          ["Brute force lockout", "Tras 10 fallos consecutivos el usuario queda BLOQUEADO 5 minutos. Permisivo para un humano, mortal para un bot."],
          ["2FA TOTP", "Opcional. Si está activo, exige código de Google Authenticator después del login."],
        ],
        [3000, 6360]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("5.2 Sesión"),
      tabla(
        ["Atributo cookie", "Valor", "Por qué"],
        [
          ["HttpOnly", "true", "JavaScript no puede leer la cookie. Si hay XSS, el atacante no roba la sesión."],
          ["Secure", "true (en producción)", "La cookie solo viaja por HTTPS. Bloquea MITM downgrade."],
          ["SameSite", "Strict", "El browser nunca envía la cookie cuando vienes de otro sitio. Mata CSRF."],
          ["Token", "32 bytes random", "Se guarda solo el SHA-256 en BD, no el token plano."],
          ["Expira", "8 horas", "Después de 8h sin actividad, hay que volver a loguearse."],
        ],
        [2200, 2000, 5160]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("5.3 Cifrado en reposo"),
      Bullet("Algoritmo: AES-256-GCM (autenticado, garantiza integridad además de confidencialidad)."),
      Bullet("Nonce aleatorio de 12 bytes por cada cifrado — ningún valor cifrado se repite, incluso si el contenido es idéntico."),
      Bullet("KEK de 32 bytes random generado al setup. Vive solo en .env del servidor."),
      Bullet("Campos cifrados en acceso_guardado: usuario_externo, password, observaciones."),
      Bullet("Campos NO cifrados (es metadata, no sensible): título, tipo, puerto, sistema_destino_id."),

      H2("5.4 Defensa en red (nginx)"),
      tabla(
        ["Header", "Qué bloquea"],
        [
          ["Strict-Transport-Security", "Si te conectaste alguna vez por HTTPS, el browser fuerza HTTPS por 1 año. Bloquea MITM downgrade."],
          ["X-Frame-Options: DENY", "Nadie puede meter tu app dentro de un iframe (clickjacking)."],
          ["X-Content-Type-Options: nosniff", "El browser no intenta adivinar tipos MIME (varios bypass de uploads cerrados)."],
          ["Referrer-Policy: strict-origin-when-cross-origin", "Al hacer click hacia un sistema externo, no le revelas la URL exacta del gestor, solo el dominio."],
          ["Permissions-Policy: geolocation/microphone/camera=()", "Aunque hubiera XSS, no podría pedir acceso a geo, micrófono o cámara."],
          ["Content-Security-Policy", "Solo scripts/estilos del propio dominio. Bloquea inyección de JS externo."],
        ],
        [3200, 6160]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("5.5 Rate limit"),
      tabla(
        ["Endpoint", "Límite", "Por qué"],
        [
          ["POST /buscar (login)", "60 req/min por IP", "Corta brute force a nivel red, complementa el lockout por usuario."],
          ["/cocina/* (autenticado)", "600 req/min por IP", "Permite uso humano (10 req/s) pero corta DoS si te roban la cookie."],
        ],
        [2800, 2200, 4360]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("5.6 Lo que aún no está blindado (asumido)"),
      Bullet("El KEK vive en .env plano. Si comprometen el usuario root del servidor, leen la KEK y descifran toda la bóveda. Para mitigar habría que usar un KMS o Vault — sobreingeniería para uso de 1 operador."),
      Bullet("Rate limit es por IP, no por usuario. Un atacante distribuido (múltiples IPs) podría intentar más logins. En la práctica, el brute force lockout por usuario (10 fallos / 5 min) cubre este caso."),

      new Paragraph({ children: [new PageBreak()] }),

      // 6. OPERACIONES
      H1("6. Operación día a día"),

      H2("6.1 Acceso a la app"),
      tabla(
        ["Tipo de acceso", "URL"],
        [
          ["Producción (publica)", "https://recetas-cocina.duckdns.org"],
          ["Login operador", "smoke@codeplex.pe (cambiar por tu correo real)"],
          ["Healthcheck", "GET https://recetas-cocina.duckdns.org/salud"],
        ],
        [3000, 6360]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("6.2 Mantenimiento del backend"),
      P(Code("pm2 logs gestor-codeplex-api"), new TextRun("    — ver logs en vivo")),
      P(Code("pm2 restart gestor-codeplex-api"), new TextRun(" — reiniciar API")),
      P(Code("pm2 status"), new TextRun("                     — ver estado de todos los procesos del server")),

      H2("6.3 Actualizar el código"),
      P("Workflow estándar:"),
      diagramaMonospace([
        "  # En local",
        "  git add . && git commit -m 'mensaje' && git push",
        "",
        "  # En el servidor (SSH)",
        "  cd /root/gestor-codeplex",
        "  git pull",
        "  # Backend Go:",
        "  cd plataforma-backend && /usr/local/go/bin/go build -o ../bin/api .",
        "  pm2 restart gestor-codeplex-api",
        "  # Frontend:",
        "  cd ../plataforma-frontend && npm run build",
        "  cp -r dist/* /var/www/recetas-cocina/",
        "  chown -R www-data:www-data /var/www/recetas-cocina",
      ]),

      H2("6.4 Si pierdes la clave del operador"),
      Bullet("No hay recuperación por email (no hay servidor SMTP configurado)."),
      Bullet("Conéctate al servidor por SSH y ejecuta cmd/bootstrap con un nuevo correo + clave temporal."),
      Bullet("Después entra al gestor con esa clave y cámbiala desde 'Cambiar contraseña'."),

      H2("6.5 Si pierdes el KEK"),
      Bullet("Toda la bóveda queda ilegible. Las credenciales cifradas no se pueden recuperar."),
      Bullet("El operador y las URLs del catálogo se conservan (no están cifrados con KEK)."),
      Bullet("Por eso: respalda el KEK en un password manager separado el día que pongas el gestor en producción."),

      new Paragraph({ children: [new PageBreak()] }),

      // 7. ENDPOINTS
      H1("7. Endpoints HTTP (referencia rápida)"),
      P("La colección Postman completa está en ", Code("postman/recetas-cocina.postman_collection.json"), ". Aquí solo el resumen."),

      H2("7.1 Públicos (sin sesión)"),
      tabla(
        ["Método", "Ruta", "Descripción"],
        [
          ["GET", "/salud", "Healthcheck. Devuelve estado del servicio y BD."],
          ["GET", "/", "Página del cebo (blog de recetas)."],
          ["POST", "/buscar", "Login disfrazado. Body: ingrediente=correo, codigo=password (form-urlencoded)."],
        ],
        [1200, 3000, 5160]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      H2("7.2 Privados (requieren sesión)"),
      tabla(
        ["Método", "Ruta", "Descripción"],
        [
          ["GET", "/cocina/identidad/perfil", "Información del operador logueado."],
          ["POST", "/cocina/identidad/cerrar-sesion", "Logout."],
          ["GET", "/cocina/identidad/sesiones", "Lista de sesiones activas del operador."],
          ["PUT", "/cocina/identidad/cambiar-password", "Cambiar clave del operador."],
          ["GET", "/cocina/sistemas", "Catálogo de URLs registradas."],
          ["POST", "/cocina/sistemas", "Registrar una URL nueva en el catálogo."],
          ["GET", "/cocina/boveda/accesos", "Lista de accesos guardados (con usuario y observaciones descifrados, sin password)."],
          ["POST", "/cocina/boveda/accesos", "Guardar un acceso (usuario y clave cifrados al recibirlos)."],
          ["PUT", "/cocina/boveda/accesos/{id}", "Editar un acceso. Si no envías 'password', conserva la actual."],
          ["DELETE", "/cocina/boveda/accesos/{id}", "Desactivar acceso (estado=REVOCADO, reversible)."],
          ["POST", "/cocina/boveda/accesos/{id}/reactivar", "Volver a poner el acceso en ACTIVO."],
          ["GET", "/cocina/boveda/accesos/{id}/autofill", "Devuelve HTML con form auto-POST al sistema externo (CSRF dependiente)."],
          ["GET", "/cocina/boveda/accesos/{id}/bookmarklet", "Devuelve JSON con usuario + clave en plano para usar en el portapapeles."],
          ["GET", "/cocina/auditoria", "Bitácora de acciones del operador."],
        ],
        [900, 3800, 4660]
      ),
      new Paragraph({ spacing: { after: 240 }, children: [] }),

      // Cierre
      H1("8. Resumen ejecutivo"),
      P(Bold("¿Es seguro?"), " Sí. Múltiples capas de defensa: HTTPS, cookie HttpOnly+Secure+SameSite, Argon2id en passwords del operador, AES-256-GCM para credenciales guardadas, brute force lockout, rate limit, headers de seguridad en nginx, anti-timing attack, 404 sigiloso, CSRF mitigado por SameSite=Strict."),
      P(Bold("¿Qué pasa si me roban la BD?"), " Las claves de los sistemas externos siguen cifradas. Sin el KEK del servidor no se pueden descifrar. Los correos de operadores y el catálogo de URLs sí quedarían expuestos (es metadata)."),
      P(Bold("¿Qué pasa si me roban el servidor entero (root)?"), " Acceso a todo: KEK + BD = bóveda descifrable. Es por eso que protejer el acceso SSH del servidor (clave privada, no password, idealmente sin contraseña reutilizada) es la línea de defensa más importante."),
      P(Bold("¿Qué le falta?"), " Un KMS externo (Vault, AWS KMS) sería el siguiente paso si el sistema crece a más operadores o si maneja secretos de mayor sensibilidad. Para uso interno de 1 operador es sobreingeniería."),

      new Paragraph({ spacing: { before: 600 }, alignment: AlignmentType.CENTER, children: [
        new TextRun({ text: "— Fin del documento —", italics: true, color: "888888" }),
      ]}),
    ],
  }],
});

const outPath = path.join(__dirname, "guia-tecnica.docx");
Packer.toBuffer(doc).then((buf) => {
  fs.writeFileSync(outPath, buf);
  console.log("OK ->", outPath, "(" + Math.round(buf.length / 1024) + " KB)");
});
