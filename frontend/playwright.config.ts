import { defineConfig, devices } from '@playwright/test';
import path from 'path';

const isCI = !!process.env.CI;

// Allow overriding the base URL via environment variable
const baseURL = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:4321';

// Output reports to a temp directory to avoid Docker volume issues
const outputDir = isCI ? './test-results' : path.join(process.env.TMPDIR || '/tmp', 'playwright-pulse');

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: isCI,
  // More retries in CI due to slower environment
  retries: isCI ? 3 : 2,
  // Fewer workers in CI to reduce contention
  workers: isCI ? 2 : 4,
  reporter: isCI
    ? [['html', { outputFolder: path.join(outputDir, 'report') }], ['github']]
    : [['html', { outputFolder: path.join(outputDir, 'report') }]],
  outputDir: path.join(outputDir, 'results'),
  // Longer timeouts for CI
  timeout: isCI ? 60000 : 45000,
  expect: {
    timeout: isCI ? 10000 : 5000,
  },
  use: {
    baseURL,
    trace: 'on-first-retry',
    actionTimeout: isCI ? 20000 : 15000,
    // Slow down actions slightly in CI
    ...(isCI && { launchOptions: { slowMo: 50 } }),
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
