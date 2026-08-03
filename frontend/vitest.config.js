import { defineConfig } from 'vitest/config'

// Dedicated test config. Deliberately NOT loading vite.config.js, so the
// vite-plugin-pwa plugin and its build steps don't run during tests.
export default defineConfig({
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.js'],
    globals: false, // import { describe, it, expect } explicitly
  },
})
