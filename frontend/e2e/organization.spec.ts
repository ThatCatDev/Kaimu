import { test, expect } from '@playwright/test';
import { setupTestEnvironment, randomId, login } from './helpers';

test.describe('Organization Management', () => {

  test('dashboard shows empty state when no organizations', async ({ page }) => {
    // Register a fresh user without creating an organization
    const testId = randomId();
    const username = `org_empty_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('No organizations')).toBeVisible();
    await expect(page.getByText('Get started by creating a new organization')).toBeVisible();
  });

  test('can navigate to create organization page from dashboard', async ({ page }) => {
    // Register a fresh user
    const testId = randomId();
    const username = `org_nav_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Click "New Organization" button
    await page.getByRole('link', { name: 'New Organization' }).first().click();

    await expect(page).toHaveURL('/organizations/new');
    await expect(page.getByRole('heading', { name: 'Create Organization' })).toBeVisible();
  });

  test('create organization form shows validation error for empty name', async ({ page }) => {
    // Register a fresh user
    const testId = randomId();
    const username = `org_valid_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    await page.goto('/organizations/new');
    await page.waitForTimeout(300);

    // Try to submit without name
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // HTML5 validation should prevent submission
    await expect(page.locator('#name')).toBeFocused();
  });

  test('can create a new organization', async ({ page }) => {
    // Register a fresh user
    const testId = randomId();
    const username = `org_create_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Test Org ${testId}`;

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    await page.goto('/organizations/new');
    await page.waitForTimeout(300);

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
    // Register a fresh user
    const testId = randomId();
    const username = `org_dash_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Dashboard Org ${testId}`;

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(300);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // Wait for redirect
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to dashboard
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Organization should be listed (in main content)
    await expect(page.getByRole('main').getByText(orgName)).toBeVisible({ timeout: 10000 });
  });

  test('can view organization detail page', async ({ page }) => {
    // Register a fresh user
    const testId = randomId();
    const username = `org_detail_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Detail Org ${testId}`;

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(300);
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
    // Register a fresh user
    const testId = randomId();
    const username = `org_empty_proj_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Empty Projects Org ${testId}`;

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(300);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();

    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Should show "No projects" message (in main content, not sidebar)
    await expect(page.getByRole('main').getByRole('heading', { name: 'No projects' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByRole('link', { name: 'New Project' })).toBeVisible();
  });

  test('can navigate to organization from dashboard', async ({ page }) => {
    // Register a fresh user
    const testId = randomId();
    const username = `org_nav_dash_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Navigate Org ${testId}`;

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an organization first
    await page.goto('/organizations/new');
    await page.waitForTimeout(300);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to dashboard
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Click on the organization card (in main content)
    await page.getByRole('main').getByText(orgName).click();

    // Should be on organization detail page
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/);
    await expect(page.getByRole('heading', { name: orgName })).toBeVisible({ timeout: 10000 });
  });

  test('cancel button on create organization returns to dashboard', async ({ page }) => {
    // Register a fresh user
    const testId = randomId();
    const username = `org_cancel_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';

    await page.goto('/register');
    await page.waitForTimeout(300);
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    await page.goto('/organizations/new');
    await page.waitForTimeout(300);

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
