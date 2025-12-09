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

test.describe('Kanban Board', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `board_e2e_${randomId}`;
  const password = 'testpassword123';
  let organizationId: string;
  let projectId: string;
  let boardId: string;
  const orgName = `Board Test Org ${randomId}`;
  const projectName = `Board Test Project ${randomId}`;
  const projectKey = `BP${randomLetters(4)}`;

  test.beforeAll(async ({ browser }) => {
    // Register a user, create an organization, and create a project
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
    const orgUrl = page.url();
    const orgMatch = orgUrl.match(/\/organizations\/([a-f0-9-]+)/);
    if (orgMatch) {
      organizationId = orgMatch[1];
    }

    // Create a project
    await page.goto(`/organizations/${organizationId}/projects/new`);
    await page.waitForTimeout(500);
    await page.fill('#name', projectName);
    await page.fill('#key', projectKey);
    await page.getByRole('button', { name: 'Create Project' }).click();

    // Extract project ID from URL
    await expect(page).toHaveURL(/\/projects\/([a-f0-9-]+)/, { timeout: 10000 });
    const projectUrl = page.url();
    const projectMatch = projectUrl.match(/\/projects\/([a-f0-9-]+)/);
    if (projectMatch) {
      projectId = projectMatch[1];
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

  test('project detail page shows kanban board link', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');

    // Should see the Kanban Board card
    await expect(page.getByRole('heading', { name: 'Kanban Board' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Default Board')).toBeVisible();
  });

  test('can navigate to kanban board from project', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');

    // Click on the Kanban Board card
    await page.getByRole('link', { name: /Kanban Board/ }).click();

    // Should be on board page
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Extract board ID
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    if (boardMatch) {
      boardId = boardMatch[1];
    }

    // Should see the board with columns
    await expect(page.getByRole('heading', { name: 'Default Board' })).toBeVisible({ timeout: 10000 });
  });

  test('kanban board shows default columns', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Should see default columns (Backlog is hidden by default)
    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'In Progress' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Done' })).toBeVisible();
  });

  test('can toggle hidden columns visibility', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Backlog should be hidden by default
    await expect(page.getByRole('heading', { name: 'Backlog' })).not.toBeVisible();

    // Toggle show hidden columns
    await page.getByLabel('Show hidden columns').click();

    // Backlog should now be visible
    await expect(page.getByRole('heading', { name: 'Backlog' })).toBeVisible({ timeout: 5000 });
  });

  test('can create a new card', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Wait for board to load
    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible({ timeout: 10000 });

    // Click add card button on the Todo column
    // Navigate up to column container (w-72) then find button
    const todoColumn = page.locator('.w-72').filter({ has: page.locator('h3:has-text("Todo")') });
    await todoColumn.getByRole('button', { name: 'Add card' }).click();

    // Fill in the card form
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
    await page.fill('#title', `Test Card ${randomId}`);
    await page.fill('#description', 'This is a test card');
    await page.selectOption('#priority', 'HIGH');
    await page.getByRole('button', { name: 'Create Card' }).click();

    // Modal should close and card should appear
    await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 5000 });
    await expect(page.getByText(`Test Card ${randomId}`)).toBeVisible({ timeout: 5000 });
  });

  test('can view and edit card details', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Wait for card from previous test
    await expect(page.getByText(`Test Card ${randomId}`)).toBeVisible({ timeout: 10000 });

    // Click on the card
    await page.getByText(`Test Card ${randomId}`).click();

    // Card detail modal should open
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#title')).toHaveValue(`Test Card ${randomId}`);

    // Update the card (auto-saves)
    await page.fill('#title', `Updated Card ${randomId}`);
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Modal should close and updated card should appear
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Old title should disappear and new title should appear
    await expect(page.getByText(`Test Card ${randomId}`)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Updated Card ${randomId}`)).toBeVisible({ timeout: 10000 });
  });

  test('can delete a card', async ({ page }) => {
    // First create a card to delete
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Create a card
    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible({ timeout: 10000 });
    const todoColumnDelete = page.locator('.w-72').filter({ has: page.locator('h3:has-text("Todo")') });
    await todoColumnDelete.getByRole('button', { name: 'Add card' }).click();
    await page.fill('#title', `Delete Me ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    // Wait for create card modal to close
    await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 5000 });
    await expect(page.getByText(`Delete Me ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Click on the card to open detail modal
    await page.getByText(`Delete Me ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Click delete button - shows confirmation modal
    await page.getByRole('button', { name: 'Delete Card' }).click();

    // Confirmation modal appears
    await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });

    // Confirm deletion - use exact match to avoid matching "Delete Card" or card title
    await page.getByRole('dialog').last().getByRole('button', { name: 'Delete', exact: true }).click();

    // Card should be gone
    await expect(page.getByText(`Delete Me ${randomId}`)).not.toBeVisible({ timeout: 5000 });
  });

  test('board breadcrumb navigation works', async ({ page }) => {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });

    // Click on Project breadcrumb
    await page.getByRole('link', { name: 'Project' }).click();
    await expect(page).toHaveURL(`/projects/${projectId}`);
  });
});
