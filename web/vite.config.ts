import path from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const gateway = env.VITE_DEV_GATEWAY_URL ?? "http://localhost:8080";

  return {
    plugins: [react()],
    resolve: { alias: { "@": path.resolve(__dirname, "./src") } },
    server: {
      port: 5173,
      proxy: {
        "/graphql": { target: gateway, changeOrigin: true },
        "/auth": { target: gateway, changeOrigin: true },
      },
    },
  };
});
