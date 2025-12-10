import { test, expect } from '@playwright/test';

// Generate a random uppercase letter string (A-Z only, for project keys)
function randomLetters(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

// Complete end-to-end user journey test
test.describe('Complete User Journey', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `journey_${randomId}`;
  const password = 'journeypassword123';

  test('full user journey: register -> create org -> create project -> logout -> login', async ({
    page,
  }) => {
    // ============================================
    // STEP 1: Register a new account
    // ============================================
    await test.step('Register new user', async () => {
      await page.goto('/register');
      await page.waitForLoadState('networkidle');

      // Fill registration form
      await page.fill('#username', testUser);
      await page.fill('#password', password);
      await page.fill('#confirmPassword', password);

      // Use Promise.all to wait for navigation while clicking
      await Promise.all([
        page.waitForURL('/', { timeout: 20000 }),
        page.getByRole('button', { name: 'Register' }).click()
      ]);
    });

    // ============================================
    // STEP 2: Navigate to dashboard
    // ============================================
    await test.step('Navigate to dashboard', async () => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText(`Welcome back, ${testUser}`)).toBeVisible();
      await expect(page.getByText('No organizations')).toBeVisible();
    });

    // ============================================
    // STEP 3: Create first organization
    // ============================================
    let orgId: string;
    const orgName = `Journey Org ${randomId}`;

    await test.step('Create first organization', async () => {
      // Click New Organization
      await page.getByRole('link', { name: 'New Organization' }).first().click();
      await expect(page).toHaveURL('/organizations/new');
      await page.waitForLoadState('networkidle');

      // Fill form
      await page.fill('#name', orgName);
      await page.fill('#description', 'My first organization in Pulse');

      // Use Promise.all to wait for navigation while clicking
      await Promise.all([
        page.waitForURL(/\/organizations\/([a-f0-9-]+)/, { timeout: 20000 }),
        page.getByRole('button', { name: 'Create Organization' }).click()
      ]);

      // Extract organization ID
      const url = page.url();
      const match = url.match(/\/organizations\/([a-f0-9-]+)/);
      if (match) {
        orgId = match[1];
      }

      // Verify organization details
      await expect(page.getByRole('heading', { name: orgName })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText('My first organization in Pulse')).toBeVisible();
    });

    // ============================================
    // STEP 4: Create first project
    // ============================================
    let projectId: string;
    const projectName = `Journey Project ${randomId}`;
    const projectKey = `JP${randomLetters(4)}`;

    await test.step('Create first project', async () => {
      // Click New Project
      await page.getByRole('link', { name: 'New Project' }).first().click();
      await expect(page).toHaveURL(`/organizations/${orgId}/projects/new`);
      await page.waitForLoadState('networkidle');
      // Wait for form to be hydrated by checking button is visible and enabled
      await expect(page.getByRole('button', { name: 'Create Project' })).toBeEnabled({ timeout: 5000 });

      // Fill form
      await page.fill('#name', projectName);
      await page.fill('#key', projectKey);
      await page.fill('#description', 'My first project in Pulse');
      await page.getByRole('button', { name: 'Create Project' }).click();

      // Should redirect to project page
      await expect(page).toHaveURL(/\/projects\/([a-f0-9-]+)/, { timeout: 10000 });

      // Extract project ID
      const url = page.url();
      const match = url.match(/\/projects\/([a-f0-9-]+)/);
      if (match) {
        projectId = match[1];
      }

      // Verify project details
      await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('main').getByText(projectKey)).toBeVisible();
    });

    // ============================================
    // STEP 5: Navigate back through the app
    // ============================================
    await test.step('Navigate back to organization', async () => {
      // Click organization link in sidebar (since breadcrumbs were removed)
      const sidebar = page.locator('aside');
      await sidebar.getByRole('link', { name: orgName }).click();
      await expect(page).toHaveURL(`/organizations/${orgId}`);

      // Verify project is listed (in main content)
      await expect(page.getByRole('main').getByText(projectName)).toBeVisible({ timeout: 10000 });
    });

    await test.step('Navigate to dashboard from organization', async () => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      // Verify organization is listed (in main content)
      await expect(page.getByRole('main').getByText(orgName)).toBeVisible({ timeout: 10000 });
    });

    // ============================================
    // STEP 6: Create second project
    // ============================================
    const project2Name = `Journey Project 2 ${randomId}`;
    const project2Key = `JPA${randomLetters(3)}`;

    await test.step('Create second project', async () => {
      // Navigate to organization (click in main content)
      await page.getByRole('main').getByText(orgName).click();
      await expect(page).toHaveURL(`/organizations/${orgId}`);
      await page.waitForLoadState('networkidle');

      // Create another project
      await page.getByRole('link', { name: 'New Project' }).first().click();
      await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+\/projects\/new/);
      await page.waitForLoadState('networkidle');

      // Wait for form to be hydrated
      await expect(page.getByRole('heading', { name: 'Create Project' })).toBeVisible({ timeout: 5000 });
      await expect(page.getByRole('button', { name: 'Create Project' })).toBeEnabled({ timeout: 5000 });

      // Fill form fields
      await page.fill('#name', project2Name);
      await page.fill('#key', project2Key);

      // Wait for form state to settle and click submit
      await page.waitForTimeout(200);

      // Use Promise.all to click and wait for navigation simultaneously
      await Promise.all([
        page.waitForURL(/\/projects\/[a-f0-9-]+/, { timeout: 15000 }),
        page.getByRole('button', { name: 'Create Project' }).click()
      ]);
    });

    await test.step('Verify both projects in organization', async () => {
      await page.goto(`/organizations/${orgId}`);
      await page.waitForLoadState('networkidle');

      await expect(page.getByRole('main').getByText(projectName)).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('main').getByText(project2Name)).toBeVisible();
    });

    // ============================================
    // STEP 7: Logout
    // ============================================
    await test.step('Logout', async () => {
      // First navigate to home page where Logout button is directly visible
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.getByRole('button', { name: 'Logout' }).click();

      // Should show login/register links
      await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('link', { name: 'Register' })).toBeVisible();
    });

    // ============================================
    // STEP 8: Login again and verify data persists
    // ============================================
    await test.step('Login again', async () => {
      await page.goto('/login');
      await page.fill('#username', testUser);
      await page.fill('#password', password);
      await page.getByRole('button', { name: 'Sign in' }).click();

      await expect(page).toHaveURL('/', { timeout: 10000 });
    });

    await test.step('Verify data persists after re-login', async () => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      // Organization should still exist (in main content)
      await expect(page.getByRole('main').getByText(orgName)).toBeVisible({ timeout: 10000 });

      // Navigate to organization and verify projects (click in main content)
      await page.getByRole('main').getByText(orgName).click();
      await expect(page.getByRole('main').getByText(projectName)).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('main').getByText(project2Name)).toBeVisible();
    });
  });

  test('user can access their projects after page refresh', async ({ page }) => {
    const user2 = `refresh_${randomId}`;
    const org2Name = `Refresh Org ${randomId}`;
    const project2Name = `Refresh Project ${randomId}`;

    // Register and create data
    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    await page.fill('#username', user2);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL('/', { timeout: 15000 });

    // Create org
    await page.goto('/organizations/new');
    await page.waitForLoadState('networkidle');
    await page.fill('#name', org2Name);
    await page.getByRole('button', { name: 'Create Organization' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 15000 });

    const orgUrl = page.url();

    // Create project
    await page.getByRole('link', { name: 'New Project' }).first().click();
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('button', { name: 'Create Project' })).toBeEnabled({ timeout: 5000 });
    await page.fill('#name', project2Name);
    await page.fill('#key', 'REFR');
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    const projectUrl = page.url();

    // Refresh the page
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Data should still be visible
    await expect(page.getByRole('heading', { name: project2Name })).toBeVisible({ timeout: 10000 });

    // Navigate to organization via refresh
    await page.goto(orgUrl);
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: org2Name })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByText(project2Name)).toBeVisible();

    // Navigate to dashboard via refresh
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('main').getByText(org2Name)).toBeVisible({ timeout: 10000 });
  });

  test('multiple organizations workflow', async ({ page }) => {
    const multiUser = `multi_${randomId}`;
    const org1 = `Multi Org 1 ${randomId}`;
    const org2 = `Multi Org 2 ${randomId}`;
    const org3 = `Multi Org 3 ${randomId}`;

    // Register
    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    await page.fill('#username', multiUser);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL('/', { timeout: 15000 });

    // Create 3 organizations
    for (const orgName of [org1, org2, org3]) {
      await page.goto('/organizations/new');
      await page.waitForLoadState('networkidle');
      await page.fill('#name', orgName);
      await page.getByRole('button', { name: 'Create Organization' }).click();
      await page.waitForLoadState('networkidle');
      await expect(page).toHaveURL(/\/organizations\/[a-f0-9-]+/, { timeout: 15000 });
    }

    // Verify all 3 appear in dashboard (in main content)
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('main').getByText(org1)).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByText(org2)).toBeVisible();
    await expect(page.getByRole('main').getByText(org3)).toBeVisible();
  });
});
