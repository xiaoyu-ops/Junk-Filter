import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    strictPort: true,
    open: !process.env.TAURI_ENV_PLATFORM,
    hmr: true,
  },
  // Tauri expects a fixed port, prevent vite from obscuring rust errors
  clearScreen: false,
  envPrefix: ['VITE_', 'TAURI_ENV_'],
  build: {
    target: 'esnext',
    minify: 'terser',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-core': ['vue', 'vue-router', 'pinia'],
        },
      },
    },
  },
})
