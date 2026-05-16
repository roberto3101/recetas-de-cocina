import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/cocina": {
        target: "http://localhost:8080",
        changeOrigin: true,
        cookieDomainRewrite: "localhost",
      },
      "/buscar": {
        target: "http://localhost:8080",
        changeOrigin: true,
        cookieDomainRewrite: "localhost",
      },
      "/salud": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/recuperacion": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
