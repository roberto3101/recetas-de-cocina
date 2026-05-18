// Genera documentos/guia-tecnica.docx
const fs = require("fs");
const path = require("path");
const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  HeadingLevel, AlignmentType, BorderStyle, WidthType, ShadingType,
  LevelFormat, PageBreak, Header, Footer, PageNumber,
} = require("docx");

const border = { style: BorderStyle.SINGLE, size: 6, color: "BBBBBB" };
const borders = { top: border, bottom: border, left: border, right: border };

const COLOR_PRIMARIO = "8B3A2A";
const COLOR_SECUNDARIO = "5C2E20";
const FILL_ZEBRA = "FAF5F0";
const FILL_HEADER = "F0DECC";
const FILL_DIAGRAMA = "F4F0EC";

function H1(t) { return new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun({ text: t, color: COLOR_PRIMARIO })] }); }
function H2(t) { return new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun({ text: t, color: COLOR_SECUNDARIO })] }); }
function P(...runs) {
  return new Paragraph({
    spacing: { after: 120 },
    children: runs.map((r) => typeof r === "string" ? new TextRun(r) : r),
  });
}
function Bold(t) { return new TextRun({ text: t, bold: true }); }
function Code(t) { return new TextRun({ text: t, font: "Consolas", color: COLOR_SECUNDARIO }); }
function Bullet(text) {
  return new Paragraph({
    numbering: { reference: "bullets", level: 0 },
    children: [new TextRun(text)],
  });
}

function cellTexto(text, opts = {}) {
  const runs = Array.isArray(text) ? text : [text];
  return new TableCell({
    borders,
    width: { size: opts.width, type: WidthType.DXA },
    shading: opts.shading ? { fill: opts.shading, type: ShadingType.CLEAR } : undefined,
    margins: { top: 80, bottom: 80, left: 120, right: 120 },
    children: runs.map((t) =>
      typeof t === "string"
        ? new Paragraph({ children: [new TextRun({ text: t, bold: !!opts.bold })] })
        : t
    ),
  });
}

function tabla(columnas, filasDatos, anchosCol) {
  const anchos = anchosCol || columnas.map(() => Math.floor(9360 / columnas.length));
  return new Table({
    width: { size: 9360, type: WidthType.DXA },
    columnWidths: anchos,
    rows: [
      new TableRow({
        tableHeader: true,
        children: columnas.map((c, i) => cellTexto(c, { width: anchos[i], shading: FILL_HEADER, bold: true })),
      }),
      ...filasDatos.map((fila, idx) =>
        new TableRow({
          children: fila.map((v, i) => cellTexto(v, { width: anchos[i], shading: idx % 2 === 0 ? FILL_ZEBRA : undefined })),
        })
      ),
    ],
  });
}

// "Diagrama" como tabla de 1 columna con párrafos en Consolas (cada paso = una fila)
function pasos(titulo, lineas) {
  const rows = lineas.map((linea, i) => new TableRow({
    children: [
      new TableCell({
        borders,
        width: { size: 9360, type: WidthType.DXA },
        shading: { fill: i % 2 === 0 ? FILL_DIAGRAMA : "FFFFFF", type: ShadingType.CLEAR },
        margins: { top: 80, bottom: 80, left: 160, right: 120 },
        children: [new Paragraph({ children: [new TextRun({ text: linea, font: "Consolas", size: 18 })] })],
      }),
    ],
  }));
  return [
    new Paragraph({ spacing: { before: 120, after: 80 }, children: [new TextRun({ text: titulo, bold: true, color: COLOR_SECUNDARIO })] }),
    new Table({ width: { size: 9360, type: WidthType.DXA }, columnWidths: [9360], rows }),
    new Paragraph({ spacing: { after: 200 }, children: [new TextRun("")] }),
  ];
}

const doc = new Document({
  creator: "roberto3101",
  title: "Guia tecnica - Gestor de accesos Codeplex",
  description: "Como se guardan los datos, las tablas, donde viven y consideraciones de seguridad",
  styles: {
    default: { document: { run: { font: "Calibri", size: 22 } } },
    paragraphStyles: [
      { id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 36, bold: true, font: "Calibri", color: COLOR_PRIMARIO },
        paragraph: { spacing: { before: 300, after: 200 }, outlineLevel: 0 } },
      { id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 28, bold: true, font: "Calibri", color: COLOR_SECUNDARIO },
        paragraph: { spacing: { before: 220, after: 140 }, outlineLevel: 1 } },
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
          children: [new TextRun({ text: "Gestor de accesos - Guia tecnica", color: "888888", size: 18 })],
        })],
      }),
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          alignment: AlignmentType.CENTER,
          children: [
            new TextRun({ text: "Pagina ", size: 18, color: "888888" }),
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
        children: [new TextRun({ text: "Guia tecnica: como se guarda y se protege la informacion", size: 28, color: "555555" })],
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "Sistemas Unificados Codeplex, 2026", italics: true, color: "888888" })],
      }),
      new Paragraph({ children: [new PageBreak()] }),

      // 1. PARA QUE SIRVE
      H1("1. Para que sirve este sistema"),
      P("El gestor de accesos es una ", Bold("boveda personal de contrasenas"), " para los sistemas internos de Codeplex. Funciona como 1Password o Bitwarden, pero corre en infraestructura propia y bajo control del operador."),
      P("Resumen en una linea: ", Bold("guarda credenciales cifradas y permite copiarlas o abrir el sistema con 1 click sin tipear la clave.")),

      H2("Que hace"),
      Bullet("Registrar URLs de sistemas a los que el operador necesita acceder (CRM, panel de control, FTP, etc.)."),
      Bullet("Guardar usuario + contrasena + observaciones para cada URL, todo cifrado en base de datos."),
      Bullet("Mostrar la lista de accesos, copiar la clave al portapapeles con un click, abrir el sistema en pestana nueva."),
      Bullet("Activar / desactivar accesos sin perder el historial (eliminacion logica)."),

      H2("Que NO hace"),
      Bullet("No se conecta a las bases de datos de los sistemas externos. No lee usuarios remotos."),
      Bullet("No envia emails de recuperacion (es uso interno, 1 operador)."),
      Bullet("No hace login automatico cross-origin (limitacion de los navegadores con sistemas que usan CSRF tipo Laravel)."),

      // 2. ARQUITECTURA
      H1("2. Arquitectura general"),
      P("El sistema tiene tres piezas que viven en el servidor de produccion:"),
      tabla(
        ["Capa", "Tecnologia", "Responsabilidad"],
        [
          ["Frontend", "React 18 + Vite + PWA, servido como estatico por nginx", "UI del operador. Form de login disfrazado de blog (cebo) y panel privado para gestionar accesos."],
          ["Reverse proxy", "nginx 1.24", "Sirve el SPA, hace proxy a /cocina, /buscar, /salud al backend Go, agrega headers de seguridad y termina TLS."],
          ["Backend", "Go 1.25, chi router, pgx driver, AES-256-GCM, Argon2id, TOTP", "Capacidades: identidad, catalogo_sistemas, boveda. Reglas de negocio, validaciones, cifrado, sesiones, auditoria."],
          ["Base de datos", "CockroachDB local en 127.0.0.1:26257, base sistemas_unificados", "6 tablas. Solo accesible desde localhost del servidor."],
          ["TLS", "Let's Encrypt, renovado por certbot cada 90 dias", "Cert valido para recetas-cocina.duckdns.org."],
        ],
        [1800, 3000, 4560]
      ),
      P(""),
      P("Todo corre en ", Code("46.225.174.213"), " (servidor Ubuntu del operador). El dominio publico es ", Code("recetas-cocina.duckdns.org"), "."),
      ...pasos("Flujo de una peticion (de afuera hacia adentro)", [
        "1) Browser del usuario  ----HTTPS---->  nginx (puerto 443)",
        "2) nginx valida TLS, agrega headers de seguridad, decide:",
        "      si la ruta es /cocina/* /buscar /salud  --> proxy al backend Go (127.0.0.1:8081)",
        "      si no, devuelve el SPA (index.html en /var/www/recetas-cocina/)",
        "3) Backend Go valida sesion (cookie sesion_cocina) y autoriza",
        "4) Backend Go consulta CockroachDB local (127.0.0.1:26257)",
        "5) Backend Go devuelve JSON -- nginx -- Browser",
      ]),

      // 3. COMO SE GUARDAN LOS DATOS
      H1("3. Como se guardan los datos"),
      P("Hay 6 tablas en la base de datos del gestor. Cada una guarda algo distinto y con distinto nivel de proteccion."),
      tabla(
        ["Tabla", "Que guarda", "Como se protege"],
        [
          ["usuario", "El operador del gestor (correo y hash de su contrasena).", "Argon2id (irreversible). Ni yo ni el operador pueden leer la clave en plano una vez guardada."],
          ["sesion_global", "Cookie de sesion activa cuando estas logueado.", "Solo se guarda el hash SHA-256 del token. Si te roban la BD no pueden suplantar sesiones."],
          ["usuario_totp", "Secreto TOTP del operador (si activa 2FA).", "AES-256-GCM cifrado con el KEK."],
          ["sistema_destino", "Catalogo de URLs a las que vas a guardar credenciales.", "Texto plano. Es metadata publica (URLs, nombre del sistema)."],
          ["acceso_guardado", "Las credenciales de la boveda: usuario y clave de cada acceso.", "Usuario, clave y observaciones cifrados AES-256-GCM. Hash HMAC-SHA256 para indice unico."],
          ["auditoria_accion", "Bitacora de cada accion que hace el operador (login, crear acceso, etc.).", "Texto plano. No registra valores sensibles, solo eventos."],
        ],
        [1872, 3600, 3888]
      ),
      P(""),

      H2("3.1 Tabla acceso_guardado en detalle"),
      P("Esta es la tabla con la informacion mas sensible. Cada fila es un acceso guardado por el operador:"),
      tabla(
        ["Columna", "Tipo", "Cifrado?", "Que contiene"],
        [
          ["id", "UUID", "no aplica", "Identificador unico de la fila"],
          ["titulo", "TEXT", "No", "Nombre que le pones al acceso, ej. 'CRM Codeplex - gerencia'"],
          ["sistema_destino_id", "UUID", "no aplica", "Referencia a la URL del catalogo"],
          ["usuario_externo", "BYTES", "Si - AES-256-GCM", "Correo o usuario del form de login externo"],
          ["usuario_externo_hash", "BYTES", "no aplica", "HMAC-SHA256 determinista para indice unico (no revela el valor)"],
          ["password_cifrada", "BYTES", "Si - AES-256-GCM", "La contrasena que se guarda para autofill"],
          ["observaciones", "BYTES", "Si - AES-256-GCM", "Notas libres del operador (PINs, preguntas de seguridad, etc.)"],
          ["tipo", "TEXT", "No", "WEB / ESCRITORIO / FTP / OTRO"],
          ["puerto", "INT2", "No", "Puerto si aplica (FTP=21, SSH=22, etc.). Opcional"],
          ["estado", "TEXT", "no aplica", "ACTIVO / REVOCADO / ELIMINADO"],
          ["creado_en, creado_por, ...", "varios", "no aplica", "Auditoria: cuando y quien"],
        ],
        [1750, 1100, 1700, 4810]
      ),
      P(""),

      H2("3.2 Por que hay un hash determinista ademas del cifrado?"),
      P("Hay un detalle tecnico interesante: AES-GCM ", Bold("no es determinista"), ". Cada vez que se cifra el mismo valor produce bytes diferentes (porque usa un nonce aleatorio). Eso es bueno para seguridad: un atacante no puede saber que dos accesos comparten el mismo usuario solo viendo los bytes."),
      P("Pero entonces, como evitamos que el operador guarde el mismo usuario dos veces para el mismo sistema? No podriamos comparar el cifrado."),
      P("Solucion: ademas del cifrado, se guarda un ", Bold("HMAC-SHA256"), " del usuario (que si es determinista). El indice unico de la tabla usa ese hash, no el valor cifrado. El hash no revela el valor pero permite comparar."),

      H2("3.3 El KEK (clave que cifra todo)"),
      P("La llave maestra que descifra el contenido de la boveda se llama ", Bold("KEK"), " (Key Encryption Key). Es una clave AES-256 de 32 bytes."),
      Bullet("Vive en el archivo /root/gestor-codeplex/.env del servidor."),
      Bullet("Permisos chmod 600: solo el usuario root del servidor puede leerla."),
      Bullet("NO se guarda en la base de datos. Si alguien dumpea la BD pero no tiene el KEK, no puede descifrar nada."),
      Bullet("Si se pierde el KEK, todas las claves cifradas se vuelven irrecuperables. Es responsabilidad del operador respaldarlo (password manager personal, por ejemplo)."),

      new Paragraph({ children: [new PageBreak()] }),

      // 4. FLUJO DE CIFRADO
      H1("4. Flujo de cifrado paso a paso"),
      ...pasos("4.1 Al guardar un acceso", [
        "1) Operador escribe en el form:    usuario='juan@x.com'",
        "                                   clave='SuperSecreta!'",
        "                                   observaciones='cuenta gerencia'",
        "2) Backend recibe el JSON          (ya viajo por HTTPS)",
        "3) Lee el KEK del .env             KEK = 32 bytes random",
        "4a) Cifra cada campo con AES-256-GCM:",
        "      usuario_cifrado   = AES(KEK, 'juan@x.com')",
        "      password_cifrada  = AES(KEK, 'SuperSecreta!')",
        "      observaciones_cif = AES(KEK, 'cuenta gerencia')",
        "4b) Hashea usuario para indice unico:",
        "      hash = HMAC-SHA256(KEK, 'juan@x.com')",
        "5) INSERT INTO acceso_guardado     (solo bytes opacos en BD)",
      ]),
      ...pasos("4.2 Al leer un acceso", [
        "1) Frontend pide GET /cocina/boveda/accesos  (con cookie de sesion)",
        "2) Backend valida sesion + 2FA               (404 sigiloso si no)",
        "3) Lee filas de acceso_guardado              (bytes cifrados)",
        "4) Por cada fila, descifra al vuelo con el KEK:",
        "      usuario       = decrypt_AES(KEK, usuario_cifrado)",
        "      observaciones = decrypt_AES(KEK, observaciones_cif)",
        "   (la clave NO se descifra en /accesos, solo en /bookmarklet)",
        "5) Devuelve JSON con valores en plano        (HTTPS al frontend)",
        "6) Frontend muestra la lista                 (clave queda como puntos)",
      ]),

      H2("4.3 Tolerancia a fallos de descifrado"),
      P("Si una fila tiene bytes que no se pueden descifrar (KEK rotada, dato corrupto, residuos de pruebas), el listado ", Bold("no falla entero"), ". Esa fila se omite con un warning en el log del backend; las demas se muestran normalmente. Esto evita que un solo dato malo bloquee toda la pantalla."),

      new Paragraph({ children: [new PageBreak()] }),

      // 5. SEGURIDAD
      H1("5. Consideraciones de seguridad"),
      P("El sistema usa varias capas de defensa. Si una falla, las otras siguen protegiendo. Lista resumida:"),

      H2("5.1 Autenticacion"),
      tabla(
        ["Defensa", "Como funciona"],
        [
          ["Argon2id en password del operador", "La clave del operador NO se guarda nunca; solo un hash Argon2id (irreversible y resistente a GPU/ASICs). Ni yo ni un atacante con la BD pueden recuperar la clave en plano."],
          ["Login disfrazado de blog", "La pagina publica parece un blog de recetas. Un atacante casual no sabe que ahi dentro hay un login. Cualquier ruta /cocina/* sin sesion devuelve 404 identico al de rutas inexistentes."],
          ["Anti-timing attack", "Si el correo NO existe, igual se ejecuta Argon2id contra un hash dummy. El tiempo de respuesta entre 'usuario existe' y 'usuario no existe' es indistinguible: bloquea enumeracion."],
          ["Brute force lockout", "Tras 10 fallos consecutivos el usuario queda BLOQUEADO 5 minutos. Permisivo para un humano, mortal para un bot."],
          ["2FA TOTP opcional", "Si esta activo, exige codigo de Google Authenticator despues del login."],
        ],
        [3000, 6360]
      ),
      P(""),

      H2("5.2 Sesion"),
      tabla(
        ["Atributo cookie", "Valor", "Por que"],
        [
          ["HttpOnly", "true", "JavaScript no puede leer la cookie. Si hay XSS, el atacante no roba la sesion."],
          ["Secure", "true en produccion", "La cookie solo viaja por HTTPS. Bloquea MITM downgrade."],
          ["SameSite", "Strict", "El browser nunca envia la cookie cuando vienes de otro sitio. Mata CSRF."],
          ["Token", "32 bytes random", "Se guarda solo el SHA-256 en BD, no el token plano."],
          ["Expira", "8 horas", "Despues de 8h sin actividad, hay que volver a loguearse."],
        ],
        [2200, 2000, 5160]
      ),
      P(""),

      H2("5.3 Cifrado en reposo"),
      Bullet("Algoritmo: AES-256-GCM (autenticado, garantiza integridad ademas de confidencialidad)."),
      Bullet("Nonce aleatorio de 12 bytes por cada cifrado: ningun valor cifrado se repite, incluso si el contenido es identico."),
      Bullet("KEK de 32 bytes random generado al setup. Vive solo en .env del servidor."),
      Bullet("Campos cifrados en acceso_guardado: usuario_externo, password, observaciones."),
      Bullet("Campos NO cifrados (es metadata, no sensible): titulo, tipo, puerto, sistema_destino_id."),

      H2("5.4 Defensa en red (nginx)"),
      tabla(
        ["Header", "Que bloquea"],
        [
          ["Strict-Transport-Security", "Si te conectaste alguna vez por HTTPS, el browser fuerza HTTPS por 1 ano. Bloquea MITM downgrade."],
          ["X-Frame-Options: DENY", "Nadie puede meter tu app dentro de un iframe (clickjacking)."],
          ["X-Content-Type-Options: nosniff", "El browser no intenta adivinar tipos MIME (varios bypass de uploads cerrados)."],
          ["Referrer-Policy", "Al hacer click hacia un sistema externo, no le revelas la URL exacta del gestor, solo el dominio."],
          ["Permissions-Policy", "Aunque hubiera XSS, no podria pedir acceso a geolocalizacion, microfono o camara."],
          ["Content-Security-Policy", "Solo scripts/estilos del propio dominio. Bloquea inyeccion de JS externo."],
        ],
        [3200, 6160]
      ),
      P(""),

      H2("5.5 Rate limit"),
      tabla(
        ["Endpoint", "Limite", "Por que"],
        [
          ["POST /buscar (login)", "60 req/min por IP", "Corta brute force a nivel red, complementa el lockout por usuario."],
          ["/cocina/* (autenticado)", "600 req/min por IP", "Permite uso humano (10 req/s) pero corta DoS si te roban la cookie."],
        ],
        [2800, 2200, 4360]
      ),
      P(""),

      H2("5.6 Lo que aun no esta blindado (asumido)"),
      Bullet("El KEK vive en .env plano. Si comprometen el usuario root del servidor, leen la KEK y descifran toda la boveda. Para mitigar habria que usar un KMS o Vault: sobreingenieria para uso de 1 operador."),
      Bullet("Rate limit es por IP, no por usuario. Un atacante distribuido (multiples IPs) podria intentar mas logins. En la practica, el brute force lockout por usuario (10 fallos / 5 min) cubre este caso."),

      new Paragraph({ children: [new PageBreak()] }),

      // 6. OPERACIONES
      H1("6. Operacion dia a dia"),

      H2("6.1 Acceso a la app"),
      tabla(
        ["Tipo de acceso", "URL"],
        [
          ["Produccion publica", "https://recetas-cocina.duckdns.org"],
          ["Login operador", "smoke@codeplex.pe (cambiar por tu correo real)"],
          ["Healthcheck", "GET https://recetas-cocina.duckdns.org/salud"],
        ],
        [3000, 6360]
      ),
      P(""),

      H2("6.2 Mantenimiento del backend"),
      P(Code("pm2 logs gestor-codeplex-api"), new TextRun("    ver logs en vivo")),
      P(Code("pm2 restart gestor-codeplex-api"), new TextRun(" reiniciar API")),
      P(Code("pm2 status"), new TextRun("                     ver estado de todos los procesos del server")),

      H2("6.3 Actualizar el codigo"),
      P("Workflow estandar:"),
      ...pasos("Pasos de despliegue", [
        "En local:",
        "  git add . && git commit -m 'mensaje' && git push",
        "",
        "En el servidor (SSH):",
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
      Bullet("No hay recuperacion por email (no hay servidor SMTP configurado)."),
      Bullet("Conectate al servidor por SSH y ejecuta cmd/bootstrap con un nuevo correo + clave temporal."),
      Bullet("Despues entra al gestor con esa clave y cambiala desde 'Cambiar contrasena'."),

      H2("6.5 Si pierdes el KEK"),
      Bullet("Toda la boveda queda ilegible. Las credenciales cifradas no se pueden recuperar."),
      Bullet("El operador y las URLs del catalogo se conservan (no estan cifrados con KEK)."),
      Bullet("Por eso: respalda el KEK en un password manager separado el dia que pongas el gestor en produccion."),

      new Paragraph({ children: [new PageBreak()] }),

      // 7. ENDPOINTS
      H1("7. Endpoints HTTP (referencia rapida)"),
      P("La coleccion Postman completa esta en ", Code("postman/recetas-cocina.postman_collection.json"), ". Aqui solo el resumen."),

      H2("7.1 Publicos (sin sesion)"),
      tabla(
        ["Metodo", "Ruta", "Descripcion"],
        [
          ["GET", "/salud", "Healthcheck. Devuelve estado del servicio y BD."],
          ["GET", "/", "Pagina del cebo (blog de recetas)."],
          ["POST", "/buscar", "Login disfrazado. Body: ingrediente=correo, codigo=password (form-urlencoded)."],
        ],
        [1200, 3000, 5160]
      ),
      P(""),

      H2("7.2 Privados (requieren sesion)"),
      tabla(
        ["Metodo", "Ruta", "Descripcion"],
        [
          ["GET", "/cocina/identidad/perfil", "Informacion del operador logueado."],
          ["POST", "/cocina/identidad/cerrar-sesion", "Logout."],
          ["GET", "/cocina/identidad/sesiones", "Lista de sesiones activas del operador."],
          ["PUT", "/cocina/identidad/cambiar-password", "Cambiar clave del operador."],
          ["GET", "/cocina/sistemas", "Catalogo de URLs registradas."],
          ["POST", "/cocina/sistemas", "Registrar una URL nueva en el catalogo."],
          ["GET", "/cocina/boveda/accesos", "Lista de accesos guardados (con usuario y observaciones descifrados, sin password)."],
          ["POST", "/cocina/boveda/accesos", "Guardar un acceso (usuario y clave cifrados al recibirlos)."],
          ["PUT", "/cocina/boveda/accesos/{id}", "Editar un acceso. Si no envias 'password', conserva la actual."],
          ["DELETE", "/cocina/boveda/accesos/{id}", "Desactivar acceso (estado=REVOCADO, reversible)."],
          ["POST", "/cocina/boveda/accesos/{id}/reactivar", "Volver a poner el acceso en ACTIVO."],
          ["GET", "/cocina/boveda/accesos/{id}/autofill", "HTML con form auto-POST al sistema externo (CSRF dependiente)."],
          ["GET", "/cocina/boveda/accesos/{id}/bookmarklet", "JSON con usuario + clave en plano para usar en el portapapeles."],
          ["GET", "/cocina/auditoria", "Bitacora de acciones del operador."],
        ],
        [900, 3800, 4660]
      ),
      P(""),

      H1("8. Resumen ejecutivo"),
      P(Bold("Es seguro?"), " Si. Multiples capas de defensa: HTTPS, cookie HttpOnly+Secure+SameSite, Argon2id en passwords del operador, AES-256-GCM para credenciales guardadas, brute force lockout, rate limit, headers de seguridad en nginx, anti-timing attack, 404 sigiloso, CSRF mitigado por SameSite=Strict."),
      P(Bold("Que pasa si me roban la BD?"), " Las claves de los sistemas externos siguen cifradas. Sin el KEK del servidor no se pueden descifrar. Los correos de operadores y el catalogo de URLs si quedarian expuestos (es metadata)."),
      P(Bold("Que pasa si me roban el servidor entero (root)?"), " Acceso a todo: KEK + BD = boveda descifrable. Es por eso que proteger el acceso SSH del servidor (clave privada, no password, idealmente sin contrasena reutilizada) es la linea de defensa mas importante."),
      P(Bold("Que le falta?"), " Un KMS externo (Vault, AWS KMS) seria el siguiente paso si el sistema crece a mas operadores o si maneja secretos de mayor sensibilidad. Para uso interno de 1 operador es sobreingenieria."),

      new Paragraph({
        spacing: { before: 600 },
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "-- Fin del documento --", italics: true, color: "888888" })],
      }),
    ],
  }],
});

const outPath = path.join(__dirname, "guia-tecnica.docx");
Packer.toBuffer(doc).then((buf) => {
  fs.writeFileSync(outPath, buf);
  console.log("OK ->", outPath, "(" + Math.round(buf.length / 1024) + " KB)");
});
