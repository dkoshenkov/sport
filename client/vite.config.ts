import fs from 'node:fs'
import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const datasetVideosDir = path.resolve(process.cwd(), '../exercises-dataset-main/videos')

function serveDatasetMedia(req: import('node:http').IncomingMessage, res: import('node:http').ServerResponse, next: () => void) {
  const requestPath = decodeURIComponent((req.url ?? '').split('?')[0])
  const fileName = path.basename(requestPath)
  if (requestPath !== `/${fileName}` || !fileName.endsWith('.gif')) {
    next()
    return
  }

  const filePath = path.join(datasetVideosDir, fileName)
  fs.stat(filePath, (error, stat) => {
    if (error || !stat.isFile()) {
      next()
      return
    }
    res.setHeader('Content-Type', 'image/gif')
    fs.createReadStream(filePath).pipe(res)
  })
}

// https://vite.dev/config/
export default defineConfig({
  publicDir: false,
  plugins: [{
    name: 'dataset-media',
    configureServer(server) {
      server.middlewares.use('/videos', serveDatasetMedia)
    },
    configurePreviewServer(server) {
      server.middlewares.use('/videos', serveDatasetMedia)
    },
  }, react(), tailwindcss()],
  server: {
    proxy: {
      '/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/healthz': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
