import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";

// VERSION-SENSITIVE (2026): TanStack Start's vite plugin generates the route
// tree from src/routes/ and wires the client/server entries. Pin the
// @tanstack/* versions on install and confirm the plugin import path
// ("@tanstack/react-start/plugin/vite") matches the installed release.
const hmrClientPort = process.env.SHIPPIN_HMR_CLIENT_PORT
  ? Number(process.env.SHIPPIN_HMR_CLIENT_PORT)
  : undefined;

export default defineConfig({
  plugins: [tanstackStart(), react()],
  server: {
    host: process.env.SHIPPIN_CONSOLE_HOST ?? "127.0.0.1",
    port: Number(process.env.SHIPPIN_CONSOLE_PORT ?? "5173"),
    allowedHosts: true,
    hmr: hmrClientPort ? { clientPort: hmrClientPort } : undefined,
    proxy: {
      "/api": { target: process.env.SHIPPIN_API_PROXY ?? "http://127.0.0.1:8081", changeOrigin: true },
      "/healthz": { target: process.env.SHIPPIN_API_PROXY ?? "http://127.0.0.1:8081", changeOrigin: true },
    },
  },
});
