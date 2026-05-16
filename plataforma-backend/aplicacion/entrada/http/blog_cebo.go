package http

import (
	"net/http"

	"sistemas-unificados/capacidades/identidad"
	"sistemas-unificados/persistencia/cockroach"
	"sistemas-unificados/plataforma/sesion"
)

const paginaInicioBlogCebo = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>Recetas del Chef — Cocina Peruana Tradicional</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  body { font-family: Georgia, serif; max-width: 820px; margin: 40px auto; padding: 0 16px; color: #333; }
  h1 { color: #8b3a2a; }
  article { border-bottom: 1px solid #eee; padding: 18px 0; }
  form.busqueda { margin: 24px 0; padding: 16px; background: #faf6f0; border-radius: 6px; }
  form.busqueda input { padding: 8px; margin-right: 8px; border: 1px solid #ddd; border-radius: 4px; }
  form.busqueda button { padding: 8px 16px; background: #8b3a2a; color: white; border: 0; border-radius: 4px; cursor: pointer; }
  footer { margin-top: 60px; color: #999; font-size: 13px; text-align: center; }
</style>
</head>
<body>
  <h1>Recetas del Chef</h1>
  <p>Bienvenido al blog de cocina peruana tradicional. Aquí compartimos recetas que han pasado de generación en generación.</p>

  <form class="busqueda" action="/buscar" method="post">
    <label>Ingrediente:&nbsp;<input type="text" name="ingrediente" autocomplete="off" required></label>
    <label>Código del chef:&nbsp;<input type="password" name="codigo" autocomplete="off" required></label>
    <button type="submit">Buscar receta</button>
  </form>

  <article>
    <h2>Ceviche de pescado</h2>
    <p>Plato bandera del Perú. Pescado fresco macerado en jugo de limón con ají limo, cebolla roja y cilantro.</p>
  </article>

  <article>
    <h2>Ají de gallina</h2>
    <p>Pollo deshilachado en crema espesa de ají amarillo, pan remojado en leche y nueces. Acompañado con papa y aceitunas.</p>
  </article>

  <article>
    <h2>Causa limeña</h2>
    <p>Puré frío de papa amarilla con ají, limón y aceite. Se rellena con pollo, atún o palta.</p>
  </article>

  <article>
    <h2>Lomo saltado</h2>
    <p>Trozos de carne salteados con cebolla, tomate y ají amarillo. Servido con arroz blanco y papas fritas.</p>
  </article>

  <footer>© Recetas del Chef — sitio personal sin fines comerciales</footer>
</body>
</html>`

const paginaBusquedaSinResultados = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>Sin resultados — Recetas del Chef</title>
<style>
  body { font-family: Georgia, serif; max-width: 820px; margin: 40px auto; padding: 0 16px; color: #333; }
  h1 { color: #8b3a2a; }
  a { color: #8b3a2a; }
</style>
</head>
<body>
  <h1>Sin resultados</h1>
  <p>No encontramos recetas con ese ingrediente. Prueba con otro término o vuelve al <a href="/">inicio</a>.</p>
</body>
</html>`

func ConstruirManejadorBlogCeboInicio() http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		ResponderRecetaInocua(escritor, http.StatusOK, paginaInicioBlogCebo)
	}
}

func ConstruirManejadorBlogCeboBuscar(conexion *cockroach.ConexionBaseDatos) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if err := peticion.ParseForm(); err != nil {
			ResponderRecetaInocua(escritor, http.StatusOK, paginaBusquedaSinResultados)
			return
		}

		correoIngresado := peticion.FormValue("ingrediente")
		passwordIngresada := peticion.FormValue("codigo")
		dispositivoIdEntrada := peticion.FormValue("dispositivo")

		resultado, err := identidad.IniciarSesion(peticion.Context(), conexion, identidad.DatosIniciarSesion{
			CorreoElectronico: correoIngresado,
			PasswordPlana:     passwordIngresada,
			DispositivoId:     dispositivoIdEntrada,
			IpOrigen:          obtenerIpRemota(peticion),
			AgenteUsuario:     peticion.UserAgent(),
		})
		if err != nil {
			ResponderRecetaInocua(escritor, http.StatusOK, paginaBusquedaSinResultados)
			return
		}

		sesion.EstablecerCookieSesion(escritor, resultado.TokenPlano, resultado.ExpiraEn)

		destino := "/panel/inventario"
		if resultado.RequiereSegundoFactor {
			destino = "/panel/verificar"
		}
		http.Redirect(escritor, peticion, destino, http.StatusSeeOther)
	}
}

const paginaInventarioPlaceholder = `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>Inventario de cocina</title></head>
<body><h1>Inventario de cocina</h1>
<p>Esta vista será reemplazada por la SPA del gestor.</p>
<p>El backend ya expone las API en /cocina/boveda/*, /cocina/sistemas/*, /cocina/auditoria.</p>
</body></html>`

const paginaVerificarPlaceholder = `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>Verificar receta</title></head>
<body><h1>Verifica tu receta</h1>
<p>Ingresa el código del segundo factor (TOTP) llamando POST /cocina/identidad/totp/validar con {"codigo":"123456"}.</p>
</body></html>`

func ConstruirManejadorCocinaInicio() http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		ResponderRecetaInocua(escritor, http.StatusOK, paginaInventarioPlaceholder)
	}
}

func ConstruirManejadorCocinaVerificar() http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		ResponderRecetaInocua(escritor, http.StatusOK, paginaVerificarPlaceholder)
	}
}
