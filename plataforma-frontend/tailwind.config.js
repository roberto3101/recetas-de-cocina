/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        cocina: {
          fondo: "#faf6f0",
          // Antes: #8b3a2a (marrón ladrillo). Ahora: naranja Tailwind 700,
          // mismo nivel de darkness/saturación pero claramente naranja.
          // Cambio global: todos los componentes que usan cocina-marron
          // (botones, títulos, focus rings, headers) pasan a naranja
          // automáticamente sin tocar JSX.
          marron: "#c2410c",
          oscuro: "#3a2418",
        },
      },
      fontFamily: {
        receta: ["Georgia", "serif"],
      },
    },
  },
  plugins: [],
};
