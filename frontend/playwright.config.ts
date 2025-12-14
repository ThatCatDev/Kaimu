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
  // Reduce retries in CI to avoid long runs - fix tests instead
  retries: isCI ? 2 : 1,
  // Use more workers in CI since tests are isolated
  workers: isCI ? 4 : 4,
  reporter: isCI
    ? [['html', { outputFolder: path.join(outputDir, 'report') }], ['github'], ['list']]
    : [['html', { outputFolder: path.join(outputDir, 'report') }]],
  outputDir: path.join(outputDir, 'results'),
  // Longer timeouts for CI
  timeout: isCI ? 90000 : 45000,
  expect: {
    timeout: isCI ? 15000 : 5000,
  },
  use: {
    baseURL,
    trace: 'on-first-retry',
    actionTimeout: isCI ? 30000 : 15000,
    navigationTimeout: isCI ? 45000 : 30000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
