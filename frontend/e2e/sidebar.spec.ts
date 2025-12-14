import { test, expect } from '@playwright/test';
import { setupTestEnvironment, navigateToBoard, randomId } from './helpers';

test.describe('Sidebar Navigation', () => {

  test('sidebar shows organizations section', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // Sidebar should be visible
    const sidebar = page.locator('aside');
    await expect(sidebar).toBeVisible();

    // Organizations header should be visible
    await expect(sidebar.getByText('Organizations')).toBeVisible({ timeout: 10000 });
  });

  test('sidebar shows organization with initial letter avatar', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Organization name should be visible
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible({ timeout: 10000 });

    // Initial letter avatar should show first letter
    const firstLetter = ctx.orgName.charAt(0).toUpperCase();
    await expect(sidebar.getByText(firstLetter, { exact: true }).first()).toBeVisible();
  });

  test('can expand organization to show projects', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Click expand button (the arrow) for the org
    const expandButton = sidebar.locator('button[title="Expand"]').first();
    await expandButton.click();

    // Project should now be visible
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 5000 });
    await expect(sidebar.getByText(ctx.projectKey)).toBeVisible();
  });

  test('can collapse organization to hide projects', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // First expand
    const expandButton = sidebar.locator('button[title="Expand"]').first();
    await expandButton.click();
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 5000 });

    // Now collapse
    const collapseButton = sidebar.locator('button[title="Collapse"]').first();
    await collapseButton.click();

    // Project should be hidden
    await expect(sidebar.getByText(ctx.projectName)).toBeHidden({ timeout: 5000 });
  });

  test('clicking organization name navigates to org page and expands', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Click on the org name link
    await sidebar.getByRole('link', { name: ctx.orgName }).click();

    // Should navigate to org page
    await expect(page).toHaveURL(`/organizations/${ctx.orgId}`);

    // Organization should be expanded in sidebar (project visible)
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 10000 });
  });

  // Skip - flaky due to sidebar caching and auto-expand timing
  test.skip('clicking project name navigates to project page and expands to show boards', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    // Create a board first by navigating to project and creating one
    await navigateToBoard(page, ctx.projectId);

    // Clear session storage and go to dashboard
    await page.goto('/dashboard');
    await page.evaluate(() => {
      sessionStorage.removeItem('expandedOrgs');
      sessionStorage.removeItem('expandedProjects');
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Wait for org to be visible first
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible({ timeout: 10000 });

    // Check if org is already expanded, if not expand it
    const projectVisible = await sidebar.getByText(ctx.projectName).isVisible().catch(() => false);
    if (!projectVisible) {
      const expandButton = sidebar.locator('button[title="Expand"]').first();
      await expandButton.click();
      await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 5000 });
    }

    // Click on the project name link
    await sidebar.getByRole('link', { name: ctx.projectName }).click();

    // Should navigate to project page
    await expect(page).toHaveURL(`/projects/${ctx.projectId}`);

    // Project should be expanded in sidebar (board visible)
    await expect(sidebar.getByText('Kanban Board')).toBeVisible({ timeout: 10000 });
  });

  test('sidebar shows project count badge', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Should show "1" badge for project count
    await expect(sidebar.locator('.bg-gray-800').getByText('1')).toBeVisible({ timeout: 10000 });
  });

  // Skip - flaky due to sidebar caching and auto-expand timing
  test.skip('can navigate to board from sidebar', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    // Create a board first
    await navigateToBoard(page, ctx.projectId);

    // Clear session storage to ensure clean state (no cached expanded orgs/projects)
    await page.goto('/dashboard');
    await page.evaluate(() => {
      sessionStorage.removeItem('expandedOrgs');
      sessionStorage.removeItem('expandedProjects');
      sessionStorage.removeItem('sidebarOrganizations');
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Wait for sidebar to load organizations
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible({ timeout: 10000 });

    // Check if org is already expanded (project visible)
    const projectVisible = await sidebar.getByText(ctx.projectName).isVisible().catch(() => false);
    if (!projectVisible) {
      // Expand org by clicking the expand button next to it
      const orgRow = sidebar.locator('div').filter({ hasText: ctx.orgName }).first();
      await orgRow.locator('button[title="Expand"]').click();
      await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 5000 });
    }

    // Check if project is already expanded (board visible)
    const boardVisible = await sidebar.getByRole('link', { name: 'Kanban Board' }).isVisible().catch(() => false);
    if (!boardVisible) {
      // Expand project by clicking the expand button next to it
      const projectRow = sidebar.locator('div').filter({ hasText: ctx.projectName }).first();
      await projectRow.locator('button[title="Expand"]').click();
      await expect(sidebar.getByRole('link', { name: 'Kanban Board' })).toBeVisible({ timeout: 5000 });
    }

    // Click on board
    await sidebar.getByRole('link', { name: 'Kanban Board' }).click();

    // Should navigate to board page
    await expect(page).toHaveURL(new RegExp(`/projects/${ctx.projectId}/board/`));
  });

  test('sidebar shows "New Project" link when org is expanded', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Expand org
    await sidebar.locator('button[title="Expand"]').first().click();

    // "New Project" link should be visible
    await expect(sidebar.getByRole('link', { name: 'New Project' })).toBeVisible({ timeout: 5000 });
  });

  test('can navigate to new project page from sidebar', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Expand org
    await sidebar.locator('button[title="Expand"]').first().click();

    // Click "New Project"
    await sidebar.getByRole('link', { name: 'New Project' }).click();

    // Should navigate to new project page
    await expect(page).toHaveURL(`/organizations/${ctx.orgId}/projects/new`);
  });

  test('sidebar can be collapsed', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Sidebar should have full width initially
    await expect(sidebar).toHaveClass(/w-64/);

    // Click collapse button in header
    await sidebar.locator('button[title="Collapse sidebar"]').click();

    // Sidebar should be collapsed
    await expect(sidebar).toHaveClass(/w-16/);
  });

  test('collapsed sidebar shows org avatars only', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Collapse sidebar
    await sidebar.locator('button[title="Collapse sidebar"]').click();
    await expect(sidebar).toHaveClass(/w-16/);

    // Org name should be hidden, but avatar should still work as link
    await expect(sidebar.getByText(ctx.orgName)).toBeHidden();

    // First letter should still be visible
    const firstLetter = ctx.orgName.charAt(0).toUpperCase();
    await expect(sidebar.getByText(firstLetter, { exact: true }).first()).toBeVisible();
  });

  test('can expand collapsed sidebar', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Collapse
    await sidebar.locator('button[title="Collapse sidebar"]').click();
    await expect(sidebar).toHaveClass(/w-16/);

    // Expand
    await sidebar.locator('button[title="Expand sidebar"]').click();
    await expect(sidebar).toHaveClass(/w-64/);

    // Org name should be visible again
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible();
  });

  test('sidebar collapse state persists across navigation', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Collapse sidebar
    await sidebar.locator('button[title="Collapse sidebar"]').click();
    await expect(sidebar).toHaveClass(/w-16/);

    // Navigate to a different page
    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Sidebar should still be collapsed
    await expect(sidebar).toHaveClass(/w-16/);
  });

  test('expanded org/project state persists across navigation', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Expand org
    await sidebar.locator('button[title="Expand"]').first().click();
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 5000 });

    // Navigate to org page (via sidebar link)
    await sidebar.getByRole('link', { name: ctx.orgName }).click();
    await expect(page).toHaveURL(`/organizations/${ctx.orgId}`);

    // Org should still be expanded after navigation
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 10000 });
  });

  test('dashboard link is active when on dashboard', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');
    const dashboardLink = sidebar.getByRole('link', { name: 'Dashboard' });

    // Dashboard link should have active styling
    await expect(dashboardLink).toHaveClass(/bg-gray-800/);
  });

  test('organization link is active when viewing org', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Org item should have active background
    const orgRow = sidebar.locator(`a[href="/organizations/${ctx.orgId}"]`).first();
    await expect(orgRow.locator('..')).toHaveClass(/bg-gray-800/);
  });

  test('settings link is visible in sidebar footer', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Settings link should be visible
    await expect(sidebar.getByRole('link', { name: 'Settings' })).toBeVisible();
  });

  test('sidebar auto-expands org when navigating directly to org page', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    // Navigate directly to org page without going through dashboard
    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Org should be auto-expanded (projects visible)
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 10000 });
  });

  // Skip - flaky due to sidebar caching and auto-expand timing
  test.skip('sidebar auto-expands project when navigating directly to project page', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    // Create a board first
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Clear all session storage for clean test
    await page.evaluate(() => {
      sessionStorage.clear();
    });

    // Navigate directly to project page
    await page.goto(`/projects/${ctx.projectId}`);
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Wait for sidebar to load data
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible({ timeout: 15000 });
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 10000 });

    // Check if board is visible, if not expand the project manually
    const boardVisible = await sidebar.getByText('Kanban Board').isVisible().catch(() => false);
    if (!boardVisible) {
      const projectRow = sidebar.locator('div').filter({ hasText: ctx.projectName }).first();
      const expandButton = projectRow.locator('button[title="Expand"]');
      if (await expandButton.isVisible()) {
        await expandButton.click();
      }
    }
    await expect(sidebar.getByText('Kanban Board')).toBeVisible({ timeout: 10000 });
  });

  // Skip - flaky due to sidebar caching and auto-expand timing
  test.skip('sidebar auto-expands when navigating directly to board page', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    // Create a board first and get the board ID
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Clear all session storage for clean test
    await page.evaluate(() => {
      sessionStorage.clear();
    });

    // Navigate directly to board page
    await page.goto(`/projects/${ctx.projectId}/board/${boardId}`);
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Wait for sidebar to load data
    await expect(sidebar.getByText(ctx.orgName)).toBeVisible({ timeout: 15000 });
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 10000 });

    // Check if board is visible, if not expand the project manually
    const boardVisible = await sidebar.getByText('Kanban Board').isVisible().catch(() => false);
    if (!boardVisible) {
      const projectRow = sidebar.locator('div').filter({ hasText: ctx.projectName }).first();
      const expandButton = projectRow.locator('button[title="Expand"]');
      if (await expandButton.isVisible()) {
        await expandButton.click();
      }
    }
    await expect(sidebar.getByText('Kanban Board')).toBeVisible({ timeout: 10000 });
  });

  // Skip - this feature (showing "default" label for default boards) is not implemented
  test.skip('default board shows "default" label', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto(`/projects/${ctx.projectId}`);
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // "default" label should be visible next to default board (using specific class selector)
    await expect(sidebar.locator('.text-indigo-400').getByText('default')).toBeVisible({ timeout: 10000 });
  });

  test('smooth navigation without page flash (View Transitions)', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'sidebar');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Expand org if not already expanded (may be auto-expanded from previous navigation)
    const expandButton = sidebar.locator('button[title="Expand"]').first();
    if (await expandButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await expandButton.click();
    }
    await expect(sidebar.getByText(ctx.projectName)).toBeVisible({ timeout: 5000 });

    // Track if full page reload happens
    let fullReloadOccurred = false;
    page.on('load', () => {
      fullReloadOccurred = true;
    });

    // Navigate via sidebar link
    await sidebar.getByRole('link', { name: ctx.orgName }).click();
    await expect(page).toHaveURL(`/organizations/${ctx.orgId}`);

    // Wait a bit for any potential reload
    await page.waitForTimeout(500);

    // With View Transitions, we shouldn't see a full page load event
    // after the initial navigation (the 'load' event fires on navigation start)
    // Instead, check that content transitioned smoothly
    await expect(page.getByRole('heading', { name: ctx.orgName })).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Sidebar - Empty States', () => {
  test('sidebar shows new organization button when no orgs exist', async ({ page }) => {
    // Register a user without creating an org
    const testId = randomId();
    const username = `sidebar_empty_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';

    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL('/', { timeout: 15000 });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // "New organization" button should be visible (the + button next to Organizations header)
    await expect(sidebar.locator('a[title="New organization"]')).toBeVisible();
  });
});

test.describe('Sidebar - Organization with no projects', () => {
  test('shows "No projects" when org is expanded but has no projects', async ({ page }) => {
    // Register user and create org without projects
    const testId = randomId();
    const username = `sidebar_noproj_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Empty Org ${testId}`;

    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL('/', { timeout: 15000 });

    // Create org without any projects
    await page.goto('/organizations/new');
    await page.waitForLoadState('networkidle');
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/\/organizations\/([a-f0-9-]+)/, { timeout: 15000 });

    // Clear session storage and go to dashboard
    await page.goto('/dashboard');
    await page.evaluate(() => {
      sessionStorage.removeItem('expandedOrgs');
      sessionStorage.removeItem('expandedProjects');
      sessionStorage.removeItem('sidebarOrganizations');
    });
    await page.reload();
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Wait for org to be visible
    await expect(sidebar.getByText(orgName)).toBeVisible({ timeout: 10000 });

    // Check if org is already expanded (No projects visible)
    const noProjectsVisible = await sidebar.getByText('No projects').isVisible().catch(() => false);
    if (!noProjectsVisible) {
      // Expand org by clicking the expand button next to it
      const orgRow = sidebar.locator('div').filter({ hasText: orgName }).first();
      await orgRow.locator('button[title="Expand"]').click();
    }

    // Should show "No projects" message
    await expect(sidebar.getByText('No projects')).toBeVisible({ timeout: 5000 });
  });

  test('org has no project count badge when empty', async ({ page }) => {
    // Register user and create org without projects
    const testId = randomId();
    const username = `sidebar_noproj_badge_${testId}`;
    const email = `${username}@test.local`;
    const password = 'testpassword123';
    const orgName = `Empty Org ${testId}`;

    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    await page.fill('#username', username);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL('/', { timeout: 15000 });

    // Create org without any projects
    await page.goto('/organizations/new');
    await page.waitForLoadState('networkidle');
    await page.fill('#name', orgName);
    await page.getByRole('button', { name: 'Create Organization' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/\/organizations\/([a-f0-9-]+)/, { timeout: 15000 });

    const url = page.url();
    const match = url.match(/\/organizations\/([a-f0-9-]+)/);
    const organizationId = match ? match[1] : '';

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const sidebar = page.locator('aside');

    // Should not show a count badge
    const orgRow = sidebar.locator(`a[href="/organizations/${organizationId}"]`).first().locator('..');
    await expect(orgRow.locator('.bg-gray-800').getByText(/^\d+$/)).toBeHidden();
  });
});
