// Genera documentos/guia-tecnica.docx
// Documento técnico explicado en lenguaje humano para que el operador del
// gestor pueda entender qué hace cada pieza y explicárselo a un tercero.
const fs = require("fs");
const path = require("path");
const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  HeadingLevel, AlignmentType, BorderStyle, WidthType, ShadingType,
  LevelFormat, PageBreak, Header, Footer, PageNumber,
} = require("docx");

const border = { style: BorderStyle.SINGLE, size: 6, color: "BBBBBB" };
const borders = { top: border, bottom: border, left: border, right: border };

// Paleta naranja para alinearse con el cambio de tema en la UI.
const COLOR_PRIMARIO = "C2410C";   // naranja oscuro (mismo que cocina-marron)
const COLOR_SECUNDARIO = "9A3412"; // naranja más oscuro para H2
const FILL_ZEBRA = "FFF7ED";       // naranja muy claro para zebra
const FILL_HEADER = "FED7AA";      // naranja medio claro para headers de tabla
const FILL_DIAGRAMA = "FEEBD3";    // beige naranja para diagramas

function H1(t) { return new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun({ text: t, color: COLOR_PRIMARIO })] }); }
function H2(t) { return new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun({ text: t, color: COLOR_SECUNDARIO })] }); }
function Bold(t) { return new TextRun({ text: t, bold: true }); }
function Code(t) { return new TextRun({ text: t, font: "Consolas", color: COLOR_SECUNDARIO }); }

// P() y Bullet() aceptan strings, runs (TextRun) o arrays mixtos. Aplanamos
// con flat() para tolerar cualquier combinación sin que el caller tenga que
// preocuparse del shape exacto. Importante porque pasar un array crudo a
// children: rompe el XML y Word no puede abrir el archivo.
function P(...args) {
  const runs = args.flat();
  return new Paragraph({
    spacing: { after: 120 },
    children: runs.map((r) => typeof r === "string" ? new TextRun(r) : r),
  });
}
function Bullet(...args) {
  const runs = args.flat();
  return new Paragraph({
    numbering: { reference: "bullets", level: 0 },
    children: runs.map((r) => typeof r === "string" ? new TextRun(r) : r),
  });
}

function cellTexto(text, opts = {}) {
  const runs = Array.isArray(text) ? text : [text];
  return new TableCell({
    borders,
    width: { size: opts.width, type: WidthType.DXA },
    shading: opts.shading ? { fill: opts.shading, type: ShadingType.CLEAR } : undefined,
    margins: { top: 80, bottom: 80, left: 120, right: 120 },
    children: runs.map((t) => {
      // string → envolver en Paragraph con TextRun
      if (typeof t === "string") {
        return new Paragraph({ children: [new TextRun({ text: t, bold: !!opts.bold })] });
      }
      // Paragraph → usar directo
      if (t instanceof Paragraph) return t;
      // TextRun u otro inline → envolverlo en Paragraph. Sin esto, docx-js
      // emite <w:r> directo dentro de <w:tc>, lo cual es XML inválido y
      // Word rechaza el archivo entero al abrirlo.
      return new Paragraph({ children: [t] });
    }),
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

// "Diagrama" como tabla de 1 columna con párrafos en Consolas
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
    new Paragraph({ spacing: { before: 100, after: 60 }, children: [new TextRun({ text: titulo, bold: true, color: COLOR_SECUNDARIO })] }),
    new Table({ width: { size: 9360, type: WidthType.DXA }, columnWidths: [9360], rows }),
    new Paragraph({ spacing: { after: 120 }, children: [new TextRun("")] }),
  ];
}

const doc = new Document({
  styles: {
    default: { document: { run: { font: "Calibri", size: 22 } } },
    paragraphStyles: [
      { id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 36, bold: true, font: "Calibri", color: COLOR_PRIMARIO },
        paragraph: { spacing: { before: 320, after: 160 }, outlineLevel: 0 } },
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
      // ============ PORTADA ============
      new Paragraph({
        spacing: { before: 2000, after: 200 },
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "Gestor de accesos", bold: true, size: 56, color: COLOR_PRIMARIO })],
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { after: 600 },
        children: [new TextRun({ text: "Guia tecnica: como funciona la boveda, que se cifra y por que", size: 28, color: "555555" })],
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "Sistemas Unificados Codeplex, 2026", italics: true, color: "888888" })],
      }),
      new Paragraph({ children: [new PageBreak()] }),

      // ============ 1. PARA QUE SIRVE ============
      H1("1. Para que sirve este sistema"),
      P("El gestor de accesos es una ", Bold("boveda personal de contrasenas"), " para los sistemas internos de Codeplex (CRM, dashpanel, correo, etc.). Funciona parecido a 1Password o Bitwarden, pero corre en infraestructura propia y bajo control del operador. Disfrazado de blog de recetas en la URL publica."),
      P("Resumen en una linea: ", Bold("guarda credenciales cifradas y permite copiarlas o abrir el sistema con un click sin tipear la clave.")),

      H2("Que hace"),
      Bullet("Guarda usuarios y contrasenas de sistemas externos, cifrados con AES-256-GCM."),
      Bullet("Permite buscar, paginar, filtrar y ordenar accesos a escala (probado con 200+ registros)."),
      Bullet("Permite copiar la clave al portapapeles o abrir el sistema directo en una pestana nueva (autofill via bookmarklet)."),
      Bullet("Soporta segundo factor TOTP (Google Authenticator / Authy) para operaciones sensibles."),
      Bullet("Soporta recuperacion de contrasena por correo electronico ('receta perdida'), con token de un solo uso que expira en 30 minutos."),
      Bullet("Audita TODA accion sensible: quien entro, que cambio, cuando, desde que IP. Imposible que algo pase sin dejar registro."),

      H2("Que NO hace"),
      Bullet("No comparte credenciales entre usuarios (cada operador ve solo lo suyo - por ahora un solo operador, pero la base ya esta lista para multi-usuario)."),
      Bullet("No sincroniza con extension de navegador (la copia es manual o via bookmarklet)."),
      Bullet("No exporta credenciales a archivo plano (decision de seguridad - ver seccion correspondiente)."),
      Bullet("No envia notificaciones de cambios. Si necesitas alertas, tienes el audit_log para auditarlas."),

      // ============ 2. ARQUITECTURA ============
      H1("2. Arquitectura general"),
      P("Tres piezas comunicandose en cadena. Cada una con un trabajo bien delimitado:"),
      tabla(
        ["Pieza", "Tecnologia", "Que hace"],
        [
          ["Frontend (navegador)", "React 18 + Vite + Tailwind + PWA", "Pantalla. Pide datos al backend, muestra formularios, copia al portapapeles. No conoce las claves."],
          ["Backend", "Go 1.25 + chi router + pgx", "Reglas de negocio. Cifra/descifra, autentica, valida, audita. Habla con la BD."],
          ["Base de datos", "CockroachDB", "Almacen. Solo ve ciphertext de las claves; no puede leerlas sin el KEK que vive en el backend."],
          ["Reverse proxy", "Nginx + Let's Encrypt", "Termina HTTPS, agrega headers de seguridad (HSTS, CSP, etc.) y pasa al backend."],
        ],
        [2400, 2400, 4560],
      ),
      ...pasos("Flujo de una peticion (de afuera hacia adentro)", [
        "Usuario en navegador",
        "  | HTTPS (TLS 1.2+)",
        "  v",
        "Nginx (puerto 443) - termina TLS, agrega headers seguridad",
        "  | HTTP plano (loopback)",
        "  v",
        "Backend Go (puerto 8081) - aplica middleware, autentica, autoriza",
        "  | TCP (loopback)",
        "  v",
        "CockroachDB (puerto 26257) - lee/escribe ciphertext",
      ]),

      // ============ 3. COMO SE GUARDAN LOS DATOS ============
      H1("3. Como se guardan los datos"),
      P("La base de datos se llama ", Code("sistemas_unificados"), " y tiene 7 tablas. Cada una con su responsabilidad concreta:"),
      tabla(
        ["Tabla", "Que guarda", "Datos cifrados?"],
        [
          ["usuario", "Operadores que pueden entrar al gestor (correo, hash de password con Argon2id, estado, contador de fallos para brute-force lockout)", "No (el password NO se guarda, se guarda su hash, que es irreversible)"],
          ["usuario_totp", "Secreto TOTP de cada operador para el segundo factor + codigos de respaldo", "SI - secreto cifrado con AES-256-GCM (clave separada del operador)"],
          ["sesion_global", "Sesiones activas de cada operador (token_hash, IP origen, agente usuario, expiracion, si paso 2FA)", "No, pero solo se guarda el hash SHA-256 del token, el plano vive solo en la cookie del navegador"],
          ["sistema_destino", "Catalogo de URLs/sistemas a los que el operador necesita entrar (CRM, dashpanel, etc.). Solo metadatos: nombre, URL, codigo - no credenciales", "No - es informacion publica de la empresa"],
          ["acceso_guardado", "El CORAZON: las credenciales en si. Un registro por cada usuario+clave que el operador guarda para un sistema_destino dado", "SI - usuario_externo, password y observaciones todos cifrados con AES-256-GCM"],
          ["auditoria_accion", "Registro inmutable de TODA accion sensible: login, cambio password, registro/edicion/eliminacion de acceso, emision de tokens de recuperacion", "No - los datos sensibles del payload no se loguean en claro"],
          ["token_recuperacion", "Tokens de un solo uso emitidos cuando alguien pide 'receta perdida'. Expiran en 30 minutos", "Solo se guarda el SHA-256 del token; el plano viaja al correo del usuario y nunca toca BD en claro"],
        ],
        [2200, 4500, 2660],
      ),

      H2("3.1 Tabla acceso_guardado en detalle"),
      P("Es la tabla que tiene la informacion mas sensible. Cada fila representa una credencial guardada (ej. tu cuenta en el CRM). Estos son los campos relevantes:"),
      tabla(
        ["Campo", "Que es", "Cifrado?"],
        [
          ["id", "UUID unico por fila", "No"],
          ["titulo", "Etiqueta humana (ej. 'CRM Codeplex - gerencia')", "No (queremos buscar por titulo, ILIKE no funciona sobre ciphertext)"],
          ["sistema_destino_id", "FK al catalogo de sistemas", "No"],
          ["usuario_externo", "Correo o user del sistema externo (ej. usuario@cliente.com)", Bold("SI - AES-256-GCM con clave 'boveda'")],
          ["usuario_externo_hash", "HMAC-SHA256 del mismo usuario. Permite buscar 'existe esta credencial?' sin descifrar (vea seccion 3.2)", "No es ciphertext, es un MAC determinista"],
          ["password_cifrada", "Clave del sistema externo", Bold("SI - AES-256-GCM con clave 'boveda'")],
          ["observaciones", "Notas libres (ej. instrucciones especiales)", Bold("SI - AES-256-GCM con clave 'boveda' (opcional, puede ser NULL)")],
          ["tipo", "WEB, ESCRITORIO, FTP, OTRO", "No (metadato simple)"],
          ["puerto", "Puerto opcional (ej. 22, 21, 443)", "No"],
          ["estado", "ACTIVO, REVOCADO o ELIMINADO", "No"],
          ["creado_en / creado_por / etc.", "Auditoria de la fila", "No"],
        ],
        [2400, 4800, 2160],
      ),

      H2("3.2 Por que hay un hash determinista ademas del cifrado?"),
      P("AES-GCM es ", Bold("no determinista"), ": el mismo texto plano cifrado dos veces da dos ciphertexts distintos (por el nonce aleatorio). Eso es bueno para seguridad pero malo para buscar duplicados."),
      P("Para evitar que el operador guarde dos veces la misma credencial (mismo usuario en el mismo sistema), guardamos en paralelo un HMAC-SHA256 del usuario. Como HMAC es determinista, mismo input = mismo hash. Eso permite:"),
      Bullet("UNIQUE INDEX sobre (sistema_destino_id, usuario_externo_hash) - rechaza duplicados sin ver el plano"),
      Bullet("Lookup rapido por usuario sin descifrar todos los registros"),
      P("El HMAC usa una clave separada (la 'clave HMAC') que tambien sale del KEK pero NO es la misma que cifra. Si alguien tuviera solo la clave HMAC podria comparar pero no descifrar - reduce el blast radius."),

      H2("3.3 El KEK (Key Encryption Key)"),
      P("Es la clave maestra. ", Bold("Vive en el .env del backend"), " (variable ", Code("KEK_GESTOR"), "), nunca en BD. Sin ella el ciphertext es inservible."),
      P("De este KEK se derivan tres claves operacionales:"),
      Bullet("Clave boveda - cifra usuario_externo, password y observaciones de acceso_guardado"),
      Bullet("Clave HMAC - genera el usuario_externo_hash determinista para el UNIQUE INDEX"),
      Bullet("Clave TOTP - cifra los secretos TOTP en usuario_totp"),
      P("Si pierdes el KEK pierdes acceso a TODAS las credenciales. Tratalo con el mismo cuidado que la cuenta root del servidor. Backup off-site recomendado."),

      H2("3.4 Tabla token_recuperacion en detalle"),
      P("Cuando alguien pide 'receta perdida' (recuperacion de contrasena), el backend:"),
      Bullet("Genera un token aleatorio de 32 bytes (256 bits) - inadivinable"),
      Bullet("Calcula SHA-256 del token"),
      Bullet("Guarda SOLO el SHA-256 en BD + usuario_id + IP + expira_en (now + 30 min)"),
      Bullet("Envia el token PLANO al correo del usuario en el link de recuperacion"),
      P("Cuando el usuario abre el link, el backend recalcula SHA-256 del token de la URL y busca el match en BD. Si encuentra una fila valida (no expirada, no consumida), permite cambiar la clave."),
      P("Despues del cambio, ", Bold("se invalidan TODAS las sesiones activas del usuario"), " - aunque alguien tuviera una sesion robada, deja de servir."),

      // ============ 4. DONDE VIVE EL CIFRADO EN EL CODIGO ============
      H1("4. Donde vive el cifrado en el codigo"),
      P("Si necesitas auditar quien cifra que, estos son los archivos exactos en el repo:"),
      tabla(
        ["Archivo", "Que hace"],
        [
          ["plataforma-backend/plataforma/cripto/cifrar_con_aes_gcm.go", "Funciones bajo nivel CifrarConAesGcm() y DescifrarConAesGcm(). Implementan AES-256-GCM con nonce aleatorio de 12 bytes prependido al ciphertext."],
          ["plataforma-backend/plataforma/cripto/claves_cifrado.go", "Lee KEK_GESTOR del entorno y deriva las 3 claves operacionales (boveda, HMAC, TOTP). Punto unico donde se centraliza la gestion del KEK."],
          ["plataforma-backend/plataforma/cripto/hashear_con_argon2id.go", "Hashea passwords de operadores con Argon2id (parametros: t=3, m=64MB, p=4, len=32). Verifica con comparacion constante en tiempo."],
          ["plataforma-backend/plataforma/cripto/hashear_token_sesion.go", "SHA-256 de tokens de sesion y de recuperacion. Determinista, rapido."],
          ["plataforma-backend/plataforma/cripto/hmac_determinista.go", "HMAC-SHA256 para el usuario_externo_hash. Permite lookup determinista sin descifrar."],
          ["plataforma-backend/capacidades/boveda/acceso_guardado.go", "Cifra usuario, password y observaciones al guardar/editar; descifra al listar y ver detalle. Centraliza toda la logica de la boveda."],
          ["plataforma-backend/capacidades/identidad/iniciar_sesion.go", "Verifica password con Argon2id. Tiene un dummy hash anti-timing: si el usuario no existe, igualmente compara contra un hash falso para que el tiempo de respuesta sea el mismo (anti-enumeracion)."],
          ["plataforma-backend/capacidades/identidad/usuario_totp.go", "Cifra/descifra el secreto TOTP con la clave dedicada. Cuando el operador activa 2FA, genera un secret de 20 bytes, lo cifra, lo guarda."],
          ["plataforma-backend/capacidades/recuperacion_password/consumir.go", "Cuando se consume un token de recovery, re-hashea la nueva password con Argon2id, marca el token como consumido (atomic CAS), invalida todas las sesiones."],
          ["plataforma-backend/capacidades/identidad/cambiar_password.go", "Cambio normal de password (con la actual). Re-hashea, invalida sesiones, audita."],
        ],
        [4800, 4560],
      ),

      // ============ 5. FLUJO DE CIFRADO PASO A PASO ============
      H1("5. Flujo de cifrado paso a paso"),
      ...pasos("5.1 Al guardar un acceso nuevo", [
        "1. Operador completa el form: titulo, sistema, usuario, clave, notas",
        "2. Frontend manda JSON a POST /cocina/boveda/accesos",
        "3. Backend valida sesion + 2FA + datos",
        "4. Backend llama a cripto.CifrarConAesGcm(claveBoveda, usuario)",
        "     -> genera nonce aleatorio 12 bytes",
        "     -> ciphertext = AES-256-GCM(usuario, nonce)",
        "     -> resultado = nonce || ciphertext || authTag",
        "5. Lo mismo para password y observaciones",
        "6. En paralelo: usuarioHash = HMAC-SHA256(claveHMAC, usuario)",
        "7. Backend ejecuta INSERT en acceso_guardado con todos los blobs cifrados + el hash determinista",
        "8. Backend escribe fila en auditoria_accion con accion='ACCESO_REGISTRADO'",
        "9. Respuesta 201 al frontend",
      ]),
      ...pasos("5.2 Al leer un acceso (autofill)", [
        "1. Operador click en titulo -> abrirEnPestana(acceso)",
        "2. Frontend hace GET /cocina/boveda/accesos/{id}/bookmarklet",
        "3. Backend lee fila de BD: ciphertext de usuario y password",
        "4. Backend descifra con cripto.DescifrarConAesGcm(claveBoveda, ciphertext)",
        "     -> verifica authTag (si no coincide -> alguien manipulo BD)",
        "     -> retorna plano",
        "5. Backend devuelve JSON {usuario, password, url_acceso} al frontend",
        "6. Frontend copia la clave al portapapeles y abre la URL en pestana nueva",
        "7. Backend escribe en auditoria_accion: accion='ACCESO_USADO'",
      ]),

      H2("5.3 Tolerancia a fallos de descifrado"),
      P("Si una fila no se puede descifrar (KEK distinta, ciphertext corrupto, etc.) el listado ", Bold("NO TUMBA la lista entera"), ". Loguea un warn ", Code("acceso_guardado.descifrado_omitido"), " con el id y la omite. Esto evita que un residuo (ej. fila de tests con otra KEK) bloquee a todo el sistema."),

      // ============ 6. RECUPERACION DE CONTRASEÑA ============
      H1("6. Recuperacion de contrasena (receta perdida)"),
      P("Flujo completo de 'olvide mi clave' disfrazado de 'receta perdida' en el blog cebo:"),
      ...pasos("6.1 Solicitar el link", [
        "1. Usuario en la landing -> link '¿Receta perdida?'",
        "2. Va a /solicitar-receta -> form con el correo",
        "3. Frontend POST /buscar/receta-perdida { ingrediente: correo }",
        "4. Backend busca el usuario por correo",
        "     - Si existe + esta ACTIVO -> sigue al paso 5",
        "     - Si NO existe / esta bloqueado / inactivo -> termina silenciosamente (anti-enum)",
        "5. Genera token aleatorio 32 bytes",
        "6. Invalida tokens previos del mismo usuario (solo el ultimo vale)",
        "7. INSERT en token_recuperacion: token_hash=SHA256(token), expira=now+30min",
        "8. Audita: accion='TOKEN_EMITIDO'",
        "9. Loguea el link en pm2 logs con nivel WARN (fallback si SMTP cae)",
        "10. Envia correo via SMTP (puerto 587 + STARTTLS, ver seccion 6.3)",
        "11. Responde al usuario el mismo mensaje generico exista o no el correo",
      ]),
      ...pasos("6.2 Consumir el link y cambiar clave", [
        "1. Usuario abre el link del correo: /receta/<token>",
        "2. Frontend hace GET /buscar/validar-receta/<token>",
        "3. Backend valida: existe?, no expirado?, no consumido?",
        "     -> si OK, devuelve correo enmascarado (ej. as****@codeplex.pe)",
        "     -> si falla, error generico 'token invalido o expirado'",
        "4. Frontend muestra form: nueva clave + confirmacion + medidor en vivo",
        "5. Usuario submitea POST /buscar/preparar-receta",
        "6. Backend valida fortaleza de nueva password",
        "7. Atomic CAS: UPDATE token_recuperacion SET consumido_en=now() WHERE id=$ AND consumido_en IS NULL AND expira_en > now()",
        "     -> si afecta 0 filas, alguien lo uso antes -> ERROR (cierra race condition)",
        "8. UPDATE usuario SET password_hash=Argon2id(nueva)",
        "9. Invalida TODAS las sesiones activas del usuario",
        "10. Desbloquea al usuario si estaba bloqueado por brute-force",
        "11. Audita: accion='PASSWORD_RESTABLECIDO'",
        "12. Frontend redirige al login con window.location.href (reload completo)",
      ]),

      H2("6.3 Envio de correo (SMTP)"),
      P("El backend usa SMTP con STARTTLS sobre puerto 587 (puerto 465 con TLS implicito esta bloqueado por el firewall del datacenter)."),
      Bullet(["Servidor: ", Code("mail.codeplex.pe"), " (variable ", Code("SMTP_HOST"), ")"]),
      Bullet(["Auth: PLAIN sobre TLS, usuario ", Code("asistencia@codeplex.pe"), " (variable ", Code("SMTP_USER"), ")"]),
      Bullet(["Timeouts: 8s para conectar + 12s para enviar. Si se pasa, abandona (no reintenta)"]),
      Bullet(["Cero reintentos por intento - el servidor SMTP bloquea la cuenta tras 5 errores seguidos"]),
      Bullet(["Circuit breaker: si fallan 3 envios consecutivos, pausa 10 minutos sin tocar SMTP (margen al limite de 5)"]),
      P("Si SMTP cae o esta pausado por circuit breaker, el link igual queda en los pm2 logs con nivel WARN. El operador puede recuperarlo y pasarselo por canal alternativo (WhatsApp, Signal). Mensaje al usuario final es el mismo - no se entera de la falla."),

      // ============ 7. BUSQUEDA, ORDENAMIENTO Y PAGINACION ============
      H1("7. Busqueda, ordenamiento y paginacion"),
      P("El listado de accesos soporta consultas a escala (probado con 200+ registros). Todo se calcula en el backend para que la busqueda funcione sobre TODA la BD, no solo sobre la pagina visible."),
      H2("7.1 Query params soportados en GET /cocina/boveda/accesos"),
      tabla(
        ["Param", "Default", "Valores validos", "Para que"],
        [
          ["limite", "50", "1..200", "Cuantas filas por respuesta. Cap a 200 si pasas mas."],
          ["offset", "0", "0..N", "Desde que posicion. Para 'pagina 3 de 40 en 40' enviarias offset=80."],
          ["orden", "creado_en", "creado_en | titulo", "Whitelist - cualquier otra cosa cae al default."],
          ["direccion", "desc", "asc | desc", "Whitelist - misma proteccion contra SQL injection."],
          ["estado", "(ambos)", "ACTIVO | REVOCADO", "Filtra por estado de la fila."],
          ["sistema_id", "(sin filtro)", "UUID valido", "Limita a un sistema especifico. UUIDs malformados se ignoran silenciosamente."],
          ["q", "(sin filtro)", "texto libre", "Busca en titulo + nombre y URL del sistema. ILIKE case-insensitive. Wildcards % y _ se escapan."],
        ],
        [1400, 1200, 2200, 4560],
      ),
      H2("7.2 Indices que aceleran las consultas"),
      P("La BD tiene 3 indices criticos creados con DDL 08_indices_busqueda.sql:"),
      Bullet([Code("idx_acceso_guardado_creado_en"), " sobre (creado_en DESC, id ASC) parcial WHERE estado != ELIMINADO - acelera el listado por defecto y la paginacion"]),
      Bullet([Code("idx_acceso_guardado_titulo_lower"), " sobre (lower(titulo) ASC, id ASC) parcial - acelera la busqueda case-insensitive y ORDER BY titulo"]),
      Bullet([Code("idx_acceso_guardado_estado_creado"), " sobre (estado, creado_en DESC, id ASC) - acelera filtros por estado + ordenamiento"]),

      H2("7.3 Limitacion: no se puede buscar usuario_externo"),
      P("La busqueda por 'q' matchea contra ", Bold("titulo"), ", ", Bold("nombre del sistema"), " y ", Bold("URL del sistema"), ". NO matchea contra usuario_externo porque ese campo esta cifrado con AES-GCM - sin descifrar todas las filas no se puede hacer substring search."),
      P("Es un trade-off de seguridad explicito: preferimos que el ciphertext quede ilegible aun sabiendo el plano de busqueda. Si en el futuro hiciera falta, se podria agregar un HMAC adicional searchable (por n-gramas) - mas complejidad."),

      // ============ 8. SEGURIDAD ============
      H1("8. Consideraciones de seguridad"),

      H2("8.1 Autenticacion (login)"),
      tabla(
        ["Mecanismo", "Detalle"],
        [
          ["Hash password", "Argon2id con t=3, m=64MB, p=4, sal aleatoria 16 bytes - resistente a GPUs y ASIC"],
          ["Comparacion", "subtle.ConstantTimeCompare - mismo tiempo aunque difiera en el primer caracter"],
          ["Anti-timing si usuario no existe", "Dummy hash en init() - compara contra el siempre, asi un atacante no puede deducir 'este correo existe' por el tiempo de respuesta"],
          ["Brute-force lockout", "10 intentos fallidos -> bloquea por 5 minutos. Configurable. Contador se resetea en login exitoso o en recovery"],
          ["Rate limit", "POST /buscar (login): 60 req/min/IP. Suficiente para humano, corta automatizacion"],
        ],
        [2400, 6960],
      ),

      H2("8.2 Sesion"),
      tabla(
        ["Aspecto", "Implementacion"],
        [
          ["Token de sesion", "32 bytes aleatorios, base64url. Solo se guarda SHA-256 en BD - el plano vive en la cookie del navegador"],
          ["Cookie", "HttpOnly + Secure + SameSite=Strict - inaccesible desde JS, solo HTTPS, no cross-site"],
          ["Expiracion", "Configurable. Por defecto sesion larga (semanas) pero invalidable desde el panel"],
          ["Invalidacion en cambio de password", "Al cambiar password (sea por recovery o flujo normal), TODAS las sesiones del usuario se marcan revocadas"],
          ["Multi-sesion", "Soportado - un usuario puede tener varios dispositivos. Cada uno tiene su token independiente"],
        ],
        [2400, 6960],
      ),

      H2("8.3 Segundo factor (TOTP)"),
      P("Las operaciones criticas requieren TOTP validado en la sesion actual:"),
      Bullet("Cambiar password"),
      Bullet("Registrar/editar/desactivar/eliminar accesos"),
      Bullet("Registrar/editar/eliminar sistemas"),
      Bullet("Registrar otro operador (multi-usuario)"),
      Bullet("Revocar TOTP"),
      P("El secreto TOTP se cifra en BD con su propia clave. Compatible con Google Authenticator, Authy, 1Password, Bitwarden y cualquier app TOTP estandar."),

      H2("8.4 Cifrado en reposo"),
      P("AES-256-GCM con nonce aleatorio por mensaje. Authenticated encryption - si alguien manipula el ciphertext en BD, al descifrar el authTag falla y la operacion aborta con error."),
      P("La clave nunca toca disco fuera del .env del backend. Si te robaran un dump de la BD ", Bold("no sirve de nada"), " sin el KEK que vive en otro sitio (.env del server)."),

      H2("8.5 Anti-enumeracion en recuperacion"),
      P("El endpoint POST /buscar/receta-perdida responde EL MISMO mensaje generico exista o no el correo. Un atacante no puede enumerar correos validos del sistema disparando recoveries."),
      P("Tambien valido para el endpoint de validacion de token: mensaje generico 'token invalido o expirado' sin diferenciar entre 'no existe', 'expiro' o 'ya consumido'."),

      H2("8.6 Defensa en red (nginx)"),
      tabla(
        ["Header / config", "Que hace"],
        [
          ["HSTS (Strict-Transport-Security)", "Fuerza HTTPS por 6 meses. Una vez que el browser lo conocio, no acepta HTTP nunca mas"],
          ["X-Frame-Options: DENY", "Bloquea que la app sea embebida en iframes - previene clickjacking"],
          ["X-Content-Type-Options: nosniff", "El browser no adivina tipos de archivo - previene MIME confusion"],
          ["Referrer-Policy: same-origin", "No filtra URLs a otros sitios"],
          ["Permissions-Policy", "Apaga geolocation, camera, microphone - cero capacidades innecesarias"],
          ["CSP (Content-Security-Policy)", "Solo carga scripts/styles del propio dominio - previene XSS si algo se escapa"],
          ["TLS 1.2+ con ciphers modernos", "Configurado en nginx, certificado Let's Encrypt con auto-renovacion"],
        ],
        [2800, 6560],
      ),

      H2("8.7 Lo que aun no esta blindado (asumido)"),
      Bullet("Multi-tenancy real: hoy hay un solo operador. La base esta lista para mas, pero no hay UI ni reglas de quien-ve-que."),
      Bullet("Backup automatico de BD: no hay job. Si el server muere, los datos se pierden. Recomendable agregar cron de cockroach BACKUP."),
      Bullet("Rotacion de KEK: no hay procedimiento de cambio de KEK con re-encryption en masa. Si la KEK se compromete, hay que rotar manualmente."),
      Bullet("Detection de anomalias: hay audit_log pero nadie lo monitorea automaticamente. Util agregar dashboards o alerts en intentos de login fuera de horario."),

      // ============ 9. OPERACION ============
      H1("9. Operacion dia a dia"),

      H2("9.1 Acceso a la app"),
      tabla(
        ["Recurso", "Donde"],
        [
          ["URL publica", Code("https://recetas-cocina.duckdns.org")],
          ["Login (correo + clave)", "Aparece como 'Buscar receta' en la landing"],
          ["Recuperacion", "Link '¿Receta perdida?' debajo del form"],
          ["Panel administrativo", Code("/panel/acceso")],
          ["Server SSH", Code("ssh root@46.225.174.213")],
          ["Alias SSH (si configurado)", Code("ssh codeplex-prod")],
        ],
        [2800, 6560],
      ),

      H2("9.2 Mantenimiento del backend (pm2)"),
      Bullet([Code("pm2 list"), " - ver procesos"]),
      Bullet([Code("pm2 logs gestor-codeplex-api"), " - ver logs en vivo"]),
      Bullet([Code("pm2 reload gestor-codeplex-api --update-env"), " - reiniciar leyendo .env nuevo"]),
      Bullet([Code("pm2 restart gestor-codeplex-api"), " - restart hard (drop conexiones)"]),

      H2("9.3 Actualizar el codigo (deploy)"),
      ...pasos("Pasos de despliegue", [
        "1. Frontend:",
        "     cd plataforma-frontend && npm run build",
        "     tar -czf - -C dist . | ssh codeplex-prod 'cd /var/www/recetas-cocina && tar -xzf -'",
        "     (opcional) borrar hashed assets viejos en /var/www/recetas-cocina/assets/",
        "",
        "2. Backend:",
        "     cd plataforma-backend && GOOS=linux GOARCH=amd64 go build -o bin/api-linux .",
        "     scp bin/api-linux codeplex-prod:/root/gestor-codeplex/bin/api.new",
        "     ssh codeplex-prod 'cd /root/gestor-codeplex/bin && mv api api.old && mv api.new api && chmod +x api && pm2 reload gestor-codeplex-api'",
        "",
        "3. DDL (si hay cambios):",
        "     ssh codeplex-prod 'cockroach sql --insecure --host=localhost:26257 --database=sistemas_unificados --file=path/al/ddl.sql'",
      ]),

      H2("9.4 Si tu jefe pierde la clave"),
      Bullet("Que entre a la landing publica"),
      Bullet("Click en '¿Receta perdida?'"),
      Bullet("Ingresa su correo (admin@codeplex.pe o el que tenga registrado)"),
      Bullet("Le llega correo desde 'Recetas del Chef <asistencia@codeplex.pe>' con link"),
      Bullet("Abre el link, escribe nueva clave dos veces, listo"),
      P("Si SMTP esta caido por cualquier razon, podes recuperar el link a mano:"),
      P([Code("ssh codeplex-prod 'pm2 logs gestor-codeplex-api --lines 100 --nostream --raw | grep link_emitido | tail -1'")]),
      P("Te muestra el link mas reciente. Se lo pasas por WhatsApp/Signal a tu jefe."),

      H2("9.5 Si pierdes el KEK"),
      P(Bold("Catastrofico. "), "Las credenciales en BD quedan inservibles - el ciphertext sigue ahi pero nadie lo puede descifrar. La unica salida es:"),
      Bullet("Si tenias backup del .env -> restaurar y seguir como estabas"),
      Bullet("Si NO tenias backup -> resetear toda la boveda y volver a ingresar manualmente cada credencial"),
      P("Recomendacion: copia el valor de KEK_GESTOR del .env en un password manager personal (1Password, Bitwarden, KeePass) y ponlo tambien en un sobre fisico cerrado guardado en otra ubicacion. Sin redundancia, un disco corrupto = perder todo."),

      H2("9.6 Probar paginacion / busqueda con datos de prueba"),
      P("Existe un seeder en /root/gestor-codeplex/bin/sembrar-test que crea 200 registros marcados:"),
      Bullet([Code("ssh codeplex-prod '/root/gestor-codeplex/bin/sembrar-test 200'"), " - crea 200 accesos con titulo '[TEST] Acceso NNNN'"]),
      P("Para borrarlos cuando termines de probar, un solo comando:"),
      P([Code("ssh codeplex-prod \"cockroach sql --insecure --host=localhost:26257 --database=sistemas_unificados <<'SQL'")]),
      P([Code("DELETE FROM acceso_guardado WHERE titulo LIKE '[TEST] Acceso %';")]),
      P([Code("DELETE FROM sistema_destino  WHERE codigo = 'test_paginacion';")]),
      P([Code("SQL\"")]),

      // ============ 10. ENDPOINTS HTTP ============
      H1("10. Endpoints HTTP (referencia rapida)"),

      H2("10.1 Publicos (sin sesion)"),
      tabla(
        ["Metodo", "Ruta", "Que hace"],
        [
          ["GET", "/", "Landing publica del blog cebo"],
          ["GET", "/salud", "Health check (status + latencia BD). Util para monitoreo."],
          ["POST", "/buscar", "Login disfrazado (form-encoded). Ingrediente=correo, codigo=password. Devuelve 303 con cookie de sesion en exito, 200 con HTML cebo en falla. Rate limit 60/min/IP."],
          ["POST", "/buscar/receta-perdida", "Solicita token de recuperacion. JSON {ingrediente: correo}. Anti-enumeracion. Rate limit 60/min/IP."],
          ["GET", "/buscar/validar-receta/{token}", "Valida si un token de recuperacion es vigente. Devuelve correo enmascarado del usuario duenio."],
          ["POST", "/buscar/preparar-receta", "Consume token y cambia password. JSON {codigo, nueva_clave}. Invalida todas las sesiones."],
        ],
        [1100, 3200, 5060],
      ),

      H2("10.2 Privados con sesion (sin requerir 2FA)"),
      tabla(
        ["Metodo", "Ruta", "Que hace"],
        [
          ["GET", "/cocina/identidad/perfil", "Devuelve datos del operador actual (correo, si paso 2FA, etc.)"],
          ["POST", "/cocina/identidad/cerrar-sesion", "Logout - invalida la cookie y la fila de sesion"],
          ["GET", "/cocina/identidad/sesiones", "Lista las sesiones activas del operador (puedes revocar otras desde aqui)"],
          ["POST", "/cocina/identidad/totp/iniciar-activacion", "Genera secret TOTP nuevo + QR para Authenticator"],
          ["POST", "/cocina/identidad/totp/confirmar-activacion", "Confirma con codigo TOTP. Si OK -> activado"],
          ["POST", "/cocina/identidad/totp/validar", "En cada sesion, antes de operaciones sensibles, valida el codigo"],
        ],
        [1100, 4200, 4060],
      ),

      H2("10.3 Privados con 2FA validado"),
      tabla(
        ["Metodo", "Ruta", "Que hace"],
        [
          ["PUT", "/cocina/identidad/cambiar-password", "Cambia password (con la actual). Invalida sesiones."],
          ["DELETE", "/cocina/identidad/sesiones/{id}", "Revoca una sesion especifica (ej. dispositivo perdido)."],
          ["POST", "/cocina/identidad/totp/revocar", "Revoca TOTP (raro - prefiere reactivar)"],
          ["POST", "/cocina/identidad/registrar-operador", "Crea un operador nuevo (para multi-usuario futuro)"],
          ["GET / POST / PUT / DELETE", "/cocina/sistemas[/{id}]", "CRUD del catalogo de sistemas externos"],
          ["GET", "/cocina/boveda/accesos", "Lista accesos paginados. Params: limite, offset, orden, direccion, estado, sistema_id, q"],
          ["POST", "/cocina/boveda/accesos", "Crea acceso (cifra usuario, password, observaciones)"],
          ["PUT", "/cocina/boveda/accesos/{id}", "Edita acceso. Si no se manda password, se conserva la anterior"],
          ["DELETE", "/cocina/boveda/accesos/{id}", "Desactiva (soft-delete -> estado REVOCADO). Reversible"],
          ["POST", "/cocina/boveda/accesos/{id}/reactivar", "Reactiva un acceso desactivado"],
          ["DELETE", "/cocina/boveda/accesos/{id}/permanente", "Borrado fisico irreversible. El frontend lo protege con modal type-to-confirm ('ELIMINAR')"],
          ["GET", "/cocina/boveda/accesos/{id}/autofill", "Devuelve credencial descifrada para autofill del bookmarklet"],
          ["GET", "/cocina/boveda/accesos/{id}/bookmarklet", "Devuelve credencial + URLs para que el frontend abra una pestana con autofill"],
          ["GET", "/cocina/auditoria", "Consulta el log de auditoria con filtros (usuario, modulo, accion, fecha)"],
        ],
        [1500, 3800, 4060],
      ),

      // ============ 11. RESUMEN EJECUTIVO ============
      H1("11. Resumen ejecutivo"),
      P("Si tuvieras que explicarle a alguien por que confiar en este sistema en una sola hoja, son estos puntos:"),
      Bullet("La password del operador se hashea con Argon2id (state-of-the-art) y se compara en tiempo constante."),
      Bullet("Las credenciales guardadas (usuario y clave de sistemas externos) se cifran con AES-256-GCM antes de tocar la BD. Sin el KEK, son ilegibles."),
      Bullet("El KEK vive solo en el .env del backend, nunca en BD. Robar un dump no sirve sin el KEK."),
      Bullet("La sesion vive en una cookie HttpOnly + Secure + SameSite=Strict. JavaScript no puede leerla."),
      Bullet("Operaciones sensibles requieren 2FA (TOTP). El secreto TOTP tambien cifrado."),
      Bullet("Recuperacion de password con token de un solo uso, expiracion 30 min, anti-enumeracion, invalida sesiones al consumir."),
      Bullet("Brute-force lockout: 10 intentos fallidos -> 5 min bloqueado. Rate limit 60 logins/min/IP."),
      Bullet("Toda accion sensible se audita en una tabla append-only. No hay forma de operar sin dejar rastro."),
      Bullet("Trafico solo HTTPS con TLS 1.2+ y headers de defensa modernos (HSTS, CSP, etc.)."),
      Bullet("Cifrado tolerante a fallos: si una fila no se puede descifrar (residuo, manipulacion), se omite con log warn en vez de tumbar la lista entera."),
      P("La filosofia es: ", Bold("defensa en profundidad."), " Si una capa cae, las otras siguen. Si te roban el dump de BD: no sirve sin KEK. Si te roban el KEK: necesitas tambien acceder a la BD. Si te robaron la cookie: caduca, no tiene 2FA, no se puede usar para operaciones sensibles."),
    ],
  }],
});

Packer.toBuffer(doc).then((buffer) => {
  const ruta = path.join(__dirname, "guia-tecnica.docx");
  fs.writeFileSync(ruta, buffer);
  console.log(`Documento generado: ${ruta} (${buffer.length} bytes)`);
});
