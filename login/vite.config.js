import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [solidPlugin()],
  base: '/login/',
  server: {
    port: 4000,
  },
  build: {
    outDir: '../.dist/frontend',
    emptyOutDir: 'true',
    target: 'esnext',
  },
});
