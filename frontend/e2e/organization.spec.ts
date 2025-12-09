import { test, expect } from '@playwright/test';

// Run tests serially to ensure clean state
test.describe.configure({ mode: 'serial' });

test.describe('Organization Management', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `org_e2e_${randomId}`;
  const password = 'testpassword123';

  test.beforeAll(async ({ browser }) => {
    // Register a user for the organization tests
    const page = await browser.newPage();
    await page.goto('/register');
    await page.waitForTimeout(500);

    await page.fill('#username', testUser);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page).toHaveURL('/', { timeout: 10000 });
    await page.close();
  });

  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.waitForTimeout(500);
    await page.fill('#username', testUser);
    await page.fill('#password', password);
    await page.getByRole('button', { name: 'Sign in' }).click();
    await expect(page.getByText(`Hello, ${testUser}`)).toBeVisible({ timeout: 10000 });
  });

  test('dashboard shows empty state when no organizations', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('No organizations')).toBeVisible();
    await expect(page.getByText('Get started by creating a new organization')).toBeVisible();
  });

  test('can navigate to create organization page from dashboard', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Click "New Organization" button
    await page.getByRole('link', { name: 'New Organization' }).first().click();

    await expect(page).toHaveURL('/organizations/new');
    await expect(page.getByRole('heading', { name: 'Create Organization' })).toBeVisible();
  });

  test('create organization form shows validation error for empty name', async ({ page }) => {
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);

    // Try to submit without name
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // HTML5 validation should prevent submission
    await expect(page.locator('#name')).toBeFocused();
  });

  test('can create a new organization', async ({ page }) => {
    const orgName = `Test Org ${randomId}`;

    await page.goto('/organizations/new');
    await page.waitForTimeout(500);

    await page.fill('#name', orgName);
    await page.fill('#description', 'A test organization for E2E testing');
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // Should redirect to organization detail page
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Organization details should be visible
    await expect(page.getByRole('heading', { name: orgName })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('A test organization for E2E testing')).toBeVisible();
  });

  test('organization appears in dashboard after creation', async ({ page }) => {
    const orgName = `Dashboard Org ${randomId}`;

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // Wait for redirect
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to dashboard
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Organization should be listed
    await expect(page.getByText(orgName)).toBeVisible({ timeout: 10000 });
  });

  test('can view organization detail page', async ({ page }) => {
    const orgName = `Detail Org ${randomId}`;

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);
    await page.fill('#name', orgName);
    await page.fill('#description', 'Organization for detail page testing');
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // Should be on organization detail page
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Verify organization details
    await expect(page.getByRole('heading', { name: orgName })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Organization for detail page testing')).toBeVisible();
  });

  test('organization detail page shows empty projects state', async ({ page }) => {
    const orgName = `Empty Projects Org ${randomId}`;

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();

    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Should show "No projects" message
    await expect(page.getByText('No projects')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: 'New Project' }).first()).toBeVisible();
  });

  test('can navigate to organization from dashboard', async ({ page }) => {
    const orgName = `Navigate Org ${randomId}`;

    // Create an organization first
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to dashboard
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Click on the organization card
    await page.getByText(orgName).click();

    // Should be on organization detail page
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/);
    await expect(page.getByRole('heading', { name: orgName })).toBeVisible({ timeout: 10000 });
  });

  test('cancel button on create organization returns to dashboard', async ({ page }) => {
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);

    await page.getByRole('link', { name: 'Cancel' }).click();

    await expect(page).toHaveURL('/dashboard');
  });

  test('unauthenticated user is redirected to login when accessing dashboard', async ({ page }) => {
    // Clear cookies to simulate logged out state
    await page.context().clearCookies();

    await page.goto('/dashboard');

    // Should be redirected to login
    await expect(page).toHaveURL('/login', { timeout: 10000 });
  });

  test('unauthenticated user is redirected when creating organization', async ({ page }) => {
    // Clear cookies
    await page.context().clearCookies();

    await page.goto('/organizations/new');
    await page.waitForLoadState('networkidle');

    // Should be redirected to login (either immediately or after form submission fails)
    // The component may handle this differently
    await page.waitForTimeout(1000);

    // Try to submit the form
    await page.fill('#name', 'Test Org');
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // Wait a bit for the API response
    await page.waitForTimeout(2000);

    // Should show error or redirect to login
    const hasUnauthorizedError = await page.getByText(/unauthorized/i).isVisible().catch(() => false);
    const hasAuthError = await page.getByText(/not authenticated/i).isVisible().catch(() => false);
    const hasError = await page.locator('.bg-red-50').isVisible().catch(() => false);
    const isLoginPage = page.url().includes('/login');

    expect(hasUnauthorizedError || hasAuthError || hasError || isLoginPage).toBeTruthy();
  });
});
