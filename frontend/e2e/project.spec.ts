import { test, expect } from '@playwright/test';

// Run tests serially to ensure clean state
test.describe.configure({ mode: 'serial' });

// Generate a random uppercase letter string (A-Z only, for project keys)
function randomLetters(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

test.describe('Project Management', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `proj_e2e_${randomId}`;
  const password = 'testpassword123';
  let organizationId: string;
  const orgName = `Project Test Org ${randomId}`;

  test.beforeAll(async ({ browser }) => {
    // Register a user and create an organization for project tests
    const page = await browser.newPage();

    // Register
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', testUser);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an organization
    await page.goto('/organizations/new');
    await page.waitForTimeout(500);
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();

    // Extract organization ID from URL
    await expect(page).toHaveURL(/\/organizations\/([a-f0-9-]+)/, { timeout: 10000 });
    const url = page.url();
    const match = url.match(/\/organizations\/([a-f0-9-]+)/);
    if (match) {
      organizationId = match[1];
    }

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

  test('can navigate to create project page from organization', async ({ page }) => {
    await page.goto(`/organizations/${organizationId}`);
    await page.waitForLoadState('networkidle');

    // Click "New Project" button (use first() since there may be multiple)
    await page.getByRole('link', { name: 'New Project' }).first().click();

    await expect(page).toHaveURL(`/organizations/${organizationId}/projects/new`);
    await expect(page.getByRole('heading', { name: 'Create Project' })).toBeVisible();
  });

  test('project key is auto-generated from name', async ({ page }) => {
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);

    await page.fill('#name', 'My Test Project');

    // Key should be auto-generated
    const keyInput = page.locator('#key');
    await expect(keyInput).toHaveValue(/MYTEST/i);
  });

  test('create project form shows validation error for empty name', async ({ page }) => {
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);

    await page.getByRole('button', { name: 'Create Project' }).click();

    // HTML5 validation should prevent submission
    await expect(page.locator('#name')).toBeFocused();
  });

  test('can create a new project', async ({ page }) => {
    const projectName = `Test Project ${randomId}`;
    const projectKey = `TP${randomLetters(4)}`;

    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);

    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.fill('#description', 'A test project for E2E testing');
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should redirect to project detail page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Project details should be visible
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(projectKey)).toBeVisible();
  });

  test('project appears in organization after creation', async ({ page }) => {
    const projectName = `Org Project ${randomId}`;
    const projectKey = `OP${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();

    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to organization detail
    await page.goto(`/organizations/${organizationId}`);
    await page.waitForLoadState('networkidle');

    // Project should be listed
    await expect(page.getByText(projectName)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(projectKey)).toBeVisible();
  });

  test('can view project detail page', async ({ page }) => {
    const projectName = `Detail Project ${randomId}`;
    const projectKey = `DP${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.fill('#description', 'Project for detail page testing');
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should be on project detail page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Verify project details
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(projectKey)).toBeVisible();
    await expect(page.getByText('Project for detail page testing')).toBeVisible();

    // Verify organization reference
    await expect(page.getByText(orgName)).toBeVisible();
  });

  test('project key must be unique within organization', async ({ page }) => {
    const projectKey = `DUP${randomLetters(3)}`;

    // Create first project
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', 'First Project');
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Try to create second project with same key
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', 'Second Project');
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should show error about duplicate key
    await expect(page.getByText(/already taken/i)).toBeVisible({ timeout: 10000 });
  });

  test('project key validation - too short', async ({ page }) => {
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);

    await page.fill('#name', 'Short Key Project');
    await page.fill('#key', 'A'); // Only 1 character
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should show validation error
    await expect(page.getByText(/2-10/i)).toBeVisible({ timeout: 10000 });
  });

  test('cancel button returns to organization page', async ({ page }) => {
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);

    await page.getByRole('link', { name: 'Cancel' }).click();

    await expect(page).toHaveURL(`/organizations/${organizationId}`);
  });

  test('can navigate to project from organization detail', async ({ page }) => {
    const projectName = `Navigate Project ${randomId}`;
    const projectKey = `NP${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to organization
    await page.goto(`/organizations/${organizationId}`);
    await page.waitForLoadState('networkidle');

    // Click on project
    await page.getByText(projectName).click();

    // Should be on project detail page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/);
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
  });

  test('project detail shows link back to organization', async ({ page }) => {
    const projectName = `Back Link Project ${randomId}`;
    const projectKey = `BL${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Organization name should be a link
    await page.getByRole('link', { name: orgName }).click();

    await expect(page).toHaveURL(`/organizations/${organizationId}`);
  });
});
