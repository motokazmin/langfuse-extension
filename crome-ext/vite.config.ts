import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { crx } from '@crxjs/vite-plugin'
import manifest from './manifest.json'

export default defineConfig({
  plugins: [
    react(),
    crx({ manifest }),
  ],
  server: {
    port: 5173,
    strictPort: true,
    cors: true,
    hmr: {
      port: 5173,
      clientPort: 5173,
    },
  },
  build: {
    outDir: 'dist',
    rollupOptions: {
      input: {
        sidepanel: 'sidepanel.html',
        'src/content-script/injector': 'src/content-script/injector.ts',
        'src/background/index': 'src/background/index.ts',
      },
      output: {
        entryFileNames: '[name].js',
      }
    },
  },
})
