import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

import { enviarFormulario } from "@/plataforma/red/cliente_api";

const recetas = [
  {
    titulo: "Ceviche de pescado",
    resumen: "Plato bandera del Perú. Pescado fresco macerado en jugo de limón con ají limo, cebolla roja y cilantro.",
  },
  {
    titulo: "Ají de gallina",
    resumen: "Pollo deshilachado en crema espesa de ají amarillo, pan remojado en leche y nueces. Acompañado con papa y aceitunas.",
  },
  {
    titulo: "Causa limeña",
    resumen: "Puré frío de papa amarilla con ají, limón y aceite. Se rellena con pollo, atún o palta.",
  },
  {
    titulo: "Lomo saltado",
    resumen: "Trozos de carne salteados con cebolla, tomate y ají amarillo. Servido con arroz blanco y papas fritas.",
  },
  {
    titulo: "Arroz con pollo",
    resumen: "Arroz teñido con cilantro y ají amarillo, presa de pollo encima. Se acompaña con salsa criolla.",
  },
];

export default function PaginaRecetas() {
  const navegar = useNavigate();
  const [enviando, setEnviando] = useState(false);
  const [ingrediente, setIngrediente] = useState("");
  const [codigo, setCodigo] = useState("");

  async function alBuscar(evento: FormEvent) {
    evento.preventDefault();
    setEnviando(true);
    try {
      await enviarFormulario("/buscar", { ingrediente, codigo });

      // Verificar si la sesión se estableció realmente consultando /perfil
      const verificacion = await fetch("/cocina/identidad/perfil", { credentials: "same-origin" });
      if (verificacion.ok) {
        const cuerpo = (await verificacion.json()) as { datos?: { segundo_factor_validado?: boolean } };
        const requiereSegundo = cuerpo.datos?.segundo_factor_validado === false;
        navegar(requiereSegundo ? "/panel/verificar" : "/panel/inventario");
        return;
      }

      navegar("/sin-resultados");
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-10">
      <h1 className="text-4xl font-bold text-cocina-marron">Recetas del Chef</h1>
      <p className="mt-2 text-cocina-oscuro">
        Bienvenido al blog de cocina peruana tradicional. Aquí compartimos recetas que han pasado de generación
        en generación.
      </p>

      <form
        onSubmit={alBuscar}
        className="my-8 rounded-md bg-cocina-fondo p-6 shadow-sm border border-amber-100"
      >
        <h2 className="text-lg font-semibold text-cocina-marron mb-4">Buscar receta</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label htmlFor="ingrediente" className="etiqueta-campo">
              Ingrediente
            </label>
            <input
              id="ingrediente"
              type="text"
              autoComplete="off"
              required
              className="campo-texto"
              value={ingrediente}
              onChange={(e) => setIngrediente(e.target.value)}
              placeholder="papa amarilla, limón, ají…"
            />
          </div>
          <div>
            <label htmlFor="codigo" className="etiqueta-campo">
              Código del chef
            </label>
            <input
              id="codigo"
              type="password"
              autoComplete="off"
              required
              className="campo-texto"
              value={codigo}
              onChange={(e) => setCodigo(e.target.value)}
            />
          </div>
        </div>
        <div className="mt-4 flex items-center justify-end">
          <button type="submit" className="boton-primario" disabled={enviando}>
            {enviando ? "Buscando…" : "Buscar receta"}
          </button>
        </div>

        {import.meta.env.DEV && (
          <div className="mt-3 border-t border-amber-200 pt-3 text-right">
            <button
              type="button"
              onClick={() => {
                setIngrediente("smoke@codeplex.pe");
                setCodigo("Smoke_Test_2026!");
              }}
              className="text-xs text-stone-500 hover:text-cocina-marron underline"
            >
              🔓 Autofill dev (smoke@codeplex.pe)
            </button>
          </div>
        )}
      </form>

      <div className="space-y-6">
        {recetas.map((receta) => (
          <article key={receta.titulo} className="border-b border-gray-200 pb-5">
            <h3 className="text-xl font-semibold text-cocina-marron">{receta.titulo}</h3>
            <p className="mt-1 text-cocina-oscuro">{receta.resumen}</p>
          </article>
        ))}
      </div>

      <footer className="mt-16 text-center text-xs text-gray-500">
        © Recetas del Chef — sitio personal sin fines comerciales
      </footer>
    </main>
  );
}
