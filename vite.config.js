import { defineConfig } from 'vite';

export default defineConfig({
  // No specific config needed for basic setup yet
  // Vite will automatically use index.html in the root as the entry point
  // and serve files from the 'public' directory.
  server: {
    proxy: {
      '/ws': {
        target: 'ws://localhost:8080', // Your Go backend
        ws: true,
      },
    },
  },
}); 