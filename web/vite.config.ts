import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The Go server (cmd/cradle, -client mode) hosts the JSON API + frame stream on
// :8090. In dev we run `vite` (HMR) and proxy every API path to it; in prod the
// built assets in dist/ are go:embed'd into the cradle binary and served at /.
const CRADLE = process.env.CRADLE_ADDR ?? 'http://localhost:8090'
const api = (p: string) => [p, { target: CRADLE, changeOrigin: true }] as const

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: Object.fromEntries(
      ['/config', '/state', '/pos', '/frame', '/walk', '/pick', '/act',
        '/chat', '/examine', '/shot', '/clip', '/sprite'].map(api),
    ),
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
