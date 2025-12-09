import { test, expect } from '@playwright/test';
import { setupTestEnvironment, randomId, randomLetters } from './helpers';

test.describe('Project Management', () => {
  test('can navigate to create project page from organization', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');

    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Click "New Project" button (use first() since there may be multiple)
    await page.getByRole('link', { name: 'New Project' }).first().click();

    await expect(page).toHaveURL(`/organizations/${ctx.orgId}/projects/new`);
    await expect(page.getByRole('heading', { name: 'Create Project' })).toBeVisible();
  });

  test('project key is auto-generated from name', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');

    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);

    await page.fill('#name', 'My Test Project');

    // Key should be auto-generated
    const keyInput = page.locator('#key');
    await expect(keyInput).toHaveValue(/MYTEST/i);
  });

  test('create project form shows validation error for empty name', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');

    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);

    await page.getByRole('button', { name: 'Create Project' }).click();

    // HTML5 validation should prevent submission
    await expect(page.locator('#name')).toBeFocused();
  });

  test('can create a new project', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');
    const testId = randomId();
    const projectName = `Test Project ${testId}`;
    const projectKey = `TP${randomLetters(4)}`;

    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);

    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.fill('#description', 'A test project for E2E testing');
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should redirect to project detail page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Project details should be visible (in main content)
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByText(projectKey)).toBeVisible();
  });

  test('project appears in organization after creation', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');
    const testId = randomId();
    const projectName = `Org Project ${testId}`;
    const projectKey = `OP${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();

    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to organization detail
    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Project should be listed (in main content)
    await expect(page.getByRole('main').getByText(projectName)).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByText(projectKey)).toBeVisible();
  });

  test('can view project detail page', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');
    const testId = randomId();
    const projectName = `Detail Project ${testId}`;
    const projectKey = `DP${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.fill('#description', 'Project for detail page testing');
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should be on project detail page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Verify project details (in main content)
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByText(projectKey)).toBeVisible();
    await expect(page.getByRole('main').getByText('Project for detail page testing')).toBeVisible();

    // Verify organization is visible in sidebar (since we removed breadcrumbs from main content)
    const sidebar = page.locator('aside');
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible();
  });

  test('project key must be unique within organization', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');
    const projectKey = `DUP${randomLetters(3)}`;

    // Create first project
    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', 'First Project');
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Try to create second project with same key
    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', 'Second Project');
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should show error about duplicate key
    await expect(page.getByText(/already taken/i)).toBeVisible({ timeout: 10000 });
  });

  test('project key validation - too short', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');

    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);

    await page.fill('#name', 'Short Key Project');
    await page.fill('#key', 'A'); // Only 1 character
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Should show validation error from backend
    await expect(page.getByText(/2-10 uppercase letters/i)).toBeVisible({ timeout: 10000 });
  });

  test('cancel button returns to organization page', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');

    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);

    await page.getByRole('link', { name: 'Cancel' }).click();

    await expect(page).toHaveURL(`/organizations/${ctx.orgId}`);
  });

  test('can navigate to project from organization detail', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');
    const testId = randomId();
    const projectName = `Navigate Project ${testId}`;
    const projectKey = `NP${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Go to organization
    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Click on project (in main content)
    await page.getByRole('main').getByText(projectName).click();

    // Should be on project detail page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/);
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible({ timeout: 10000 });
  });

  test('project detail shows link back to organization', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'proj');
    const testId = randomId();
    const projectName = `Back Link Project ${testId}`;
    const projectKey = `BL${randomLetters(4)}`;

    // Create a project
    await page.goto(`/organizations/${ctx.orgId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });

    // Organization name should be a link (in main content)
    await page.getByRole('main').getByRole('link', { name: ctx.orgName }).click();

    await expect(page).toHaveURL(`/organizations/${ctx.orgId}`);
  });
});
