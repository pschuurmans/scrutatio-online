import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  // Relative base so the build works on both a root domain and a GitHub Pages
  // subpath (https://<user>.github.io/<repo>/). Data is fetched with relative
  // paths from bibleData.js, so this keeps asset + SW URLs correct either way.
  base: './',
  plugins: [
    svelte(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['vite.svg', 'icons/*.png'],
      manifest: {
        name: 'Bijbel API',
        short_name: 'Bijbel',
        description: 'Bestudeer de Bijbel met 340.000+ kruisverwijzingen. Klik op een vers en ontdek direct gerelateerde passages. ',
        theme_color: '#4A90E2',
        background_color: '#ffffff',
        display: 'standalone',
        icons: [
          {
            src: './icons/icon-192x192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'any maskable'
          },
          {
            src: './icons/icon-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        navigateFallback: null,
        runtimeCaching: [
          {
            // Static Bible data (books + cross references) is immutable, so
            // cache it aggressively once downloaded (lazy-loaded per book).
            urlPattern: /\/data\/.*\.json$/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'bijbel-data',
              expiration: {
                maxEntries: 200,
                maxAgeSeconds: 60 * 60 * 24 * 365 // 1 year (data is static)
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          }
        ]
      },
      devOptions: {
        enabled: false
      }
    })
  ],
  build: {
    outDir: 'dist',
    sourcemap: false
  }
})
