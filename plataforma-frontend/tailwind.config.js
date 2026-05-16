/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        cocina: {
          fondo: "#faf6f0",
          marron: "#8b3a2a",
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
