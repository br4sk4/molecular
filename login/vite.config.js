import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [solidPlugin()],
  base: '/login/ui/',
  server: {
    port: 4000,
  },
  build: {
    outDir: '../.dist/frontend',
    emptyOutDir: 'true',
    target: 'esnext',
  },
});
