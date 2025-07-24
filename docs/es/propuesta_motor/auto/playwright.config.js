import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'list',
  timeout: 120000, // 2 minutos de timeout global
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'off',
    actionTimeout: 10000,
    navigationTimeout: 30000,
  },

  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        viewport: { width: 1366, height: 768 },
        launchOptions: {
          slowMo: 1000, // Más lento para que sea visible
          headless: false,
          args: ['--start-maximized']
        },
      },
    },
    {
      name: 'firefox',
      use: { 
        ...devices['Desktop Firefox'],
        viewport: { width: 1366, height: 768 },
      },
    },
  ],

  webServer: {
    command: 'echo "Using existing server"',
    port: 3000,
    timeout: 120 * 1000,
    reuseExistingServer: true,
  },
});