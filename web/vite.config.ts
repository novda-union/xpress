import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const proxyTarget = env.VITE_PROXY_TARGET || 'http://localhost:8080'

  const proxyPaths = ['/auth', '/discover', '/branches', '/stores', '/orders', '/health', '/ws']
  const proxy = Object.fromEntries(
    proxyPaths.map((pathName) => [
      pathName,
      {
        target: proxyTarget,
        changeOrigin: true,
        ws: pathName === '/ws',
      },
    ]),
  )

  return {
    plugins: [react(), tailwindcss()],
    build: {
      chunkSizeWarningLimit: 1100,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules')) {
              return
            }

            if (id.includes('react-router-dom')) {
              return 'router'
            }

            if (id.includes('maplibre-gl')) {
              return 'map'
            }

            if (id.includes('@telegram-apps/sdk')) {
              return 'telegram'
            }

            if (id.includes('lucide-react')) {
              return 'icons'
            }

            if (id.includes('react') || id.includes('scheduler')) {
              return 'react-vendor'
            }

            if (id.includes('zustand')) {
              return 'state'
            }
          },
        },
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      allowedHosts: true,
      proxy,
    },
  }
})
