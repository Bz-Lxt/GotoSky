import { defineConfig } from '@playwright/test'
export default defineConfig({
  testDir: '.',
  timeout: 60000,
  use: { baseURL: process.env.E2E_BASE || 'http://127.0.0.1:28353', viewport: { width: 1440, height: 900 } },
})
