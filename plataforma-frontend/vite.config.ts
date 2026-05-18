import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";
import path from "path";
import { webcrypto } from "crypto";

// Node 18 no expone crypto como global; vite-plugin-pwa lo necesita en build
if (!(globalThis as unknown as { crypto?: Crypto }).crypto) {
  (globalThis as unknown as { crypto: Crypto }).crypto = webcrypto as unknown as Crypto;
}

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["favicon.svg", "apple-touch-icon.png"],
      manifest: {
        name: "Recetas del Chef",
        short_name: "Recetas",
        description: "Gestor de accesos",
        theme_color: "#8b3a2a",
        background_color: "#f5f5f4",
        display: "standalone",
        orientation: "portrait",
        scope: "/",
        start_url: "/",
        icons: [
          { src: "icon-192.png", sizes: "192x192", type: "image/png" },
          { src: "icon-512.png", sizes: "512x512", type: "image/png" },
          { src: "icon-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
        ],
      },
      workbox: {
        // Las rutas del backend NUNCA se cachean (siempre red)
        navigateFallbackDenylist: [/^\/cocina/, /^\/buscar/, /^\/salud/, /^\/recuperacion/],
        runtimeCaching: [
          {
            urlPattern: /^\/cocina\/.*/i,
            handler: "NetworkOnly",
          },
        ],
      },
    }),
  ],
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
    },
  },
});
