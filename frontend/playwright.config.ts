import { defineConfig, devices } from '@playwright/test';
import path from 'path';

// Output reports to a temp directory to avoid Docker volume issues
const outputDir = process.env.CI ? './test-results' : path.join(process.env.TMPDIR || '/tmp', 'playwright-pulse');

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [['html', { outputFolder: path.join(outputDir, 'report') }]],
  outputDir: path.join(outputDir, 'results'),
  use: {
    baseURL: 'http://localhost:4321',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // Note: Start docker compose before running tests
  // webServer is disabled since we expect Docker to be running
});
