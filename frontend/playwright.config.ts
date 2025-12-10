import { defineConfig, devices } from '@playwright/test';
import path from 'path';

// Output reports to a temp directory to avoid Docker volume issues
const outputDir = process.env.CI ? './test-results' : path.join(process.env.TMPDIR || '/tmp', 'playwright-pulse');

export default defineConfig({
  testDir: './e2e',
  // Run test files in parallel, but tests within a file run serially (via describe.configure)
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  // Add retries to handle transient failures
  retries: process.env.CI ? 6 : 6,
  // Use more workers for faster test execution
  workers: process.env.CI ? 4 : 8,
  reporter: [['html', { outputFolder: path.join(outputDir, 'report') }]],
  outputDir: path.join(outputDir, 'results'),
  // Increase timeout for slower operations
  timeout: 45000,
  use: {
    baseURL: 'http://localhost:4321',
    trace: 'on-first-retry',
    // Add action timeout
    actionTimeout: 15000,
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
