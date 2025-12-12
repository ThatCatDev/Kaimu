import { test, expect, type Page } from '@playwright/test';
import { setupTestEnvironment, navigateToBoard, createCard, type TestContext } from './helpers';

/**
 * Helper to navigate to the planning page
 */
async function navigateToPlanning(page: Page, projectId: string, boardId: string) {
  await page.goto(`/projects/${projectId}/board/${boardId}/planning`);
  await page.waitForLoadState('networkidle');
  // Wait for the planning view to load - should see Backlog section
  await expect(page.getByText('Backlog', { exact: true }).first()).toBeVisible({ timeout: 10000 });
}

/**
 * Helper to open the sprint selector dropdown
 */
async function openSprintSelector(page: Page) {
  const trigger = page.locator('button[data-popover-trigger]').first();
  await expect(trigger).toBeVisible({ timeout: 10000 });
  await trigger.click();
  await expect(page.getByRole('heading', { name: 'Sprints' })).toBeVisible({ timeout: 5000 });
}

/**
 * Helper to create a sprint via the UI
 */
async function createSprint(page: Page, name: string) {
  await openSprintSelector(page);
  await page.locator('button').filter({ hasText: 'Create Sprint' }).click();
  await expect(page.getByRole('dialog', { name: 'Create Sprint' })).toBeVisible({ timeout: 5000 });
  const modal = page.getByRole('dialog', { name: 'Create Sprint' });
  await modal.getByPlaceholder('e.g., Sprint 1').fill(name);
  await modal.getByRole('button', { name: 'Create Sprint' }).click();
  await expect(modal).not.toBeVisible({ timeout: 5000 });
}

/**
 * Helper to start a sprint
 */
async function startSprint(page: Page, sprintName: string) {
  await openSprintSelector(page);
  const popover = page.locator('[data-popover-content]');
  const sprintRow = popover.locator('.px-3.py-2').filter({
    has: page.getByRole('button', { name: sprintName })
  });
  const startButton = sprintRow.getByRole('button', { name: 'Start' });
  await expect(startButton).toBeVisible({ timeout: 5000 });
  await startButton.click();
  await expect(page.getByText(/Started/)).toBeVisible({ timeout: 10000 });
  await page.waitForTimeout(500);
}

test.describe('Tab Navigation', () => {

  test('can see all three tabs on board page', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tab');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Should see all three tabs
    await expect(page.getByRole('link', { name: 'Board' }).or(page.locator('span').filter({ hasText: 'Board' }).first())).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('link', { name: 'Planning' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('link', { name: 'Metrics' })).toBeVisible({ timeout: 5000 });
  });

  test('can navigate from Board to Planning tab', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tab');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Click Planning tab
    await page.getByRole('link', { name: 'Planning' }).click();

    // Should navigate to planning page
    await expect(page).toHaveURL(`/projects/${ctx.projectId}/board/${actualBoardId}/planning`, { timeout: 10000 });

    // Should see planning content - Backlog section
    await expect(page.getByText('Backlog', { exact: true }).first()).toBeVisible({ timeout: 10000 });
  });

  test('can navigate from Board to Metrics tab', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tab');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Click Metrics tab
    await page.getByRole('link', { name: 'Metrics' }).click();

    // Should navigate to metrics page
    await expect(page).toHaveURL(`/projects/${ctx.projectId}/board/${actualBoardId}/metrics`, { timeout: 10000 });

    // Should see metrics content
    await expect(page.getByText('Backlog', { exact: true }).first()).toBeVisible({ timeout: 10000 });
  });

  test('can navigate from Planning to Board tab', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tab');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Go to planning first
    await page.getByRole('link', { name: 'Planning' }).click();
    await expect(page).toHaveURL(/\/planning/, { timeout: 10000 });

    // Now click Board tab to go back
    await page.getByRole('link', { name: 'Board' }).click();

    // Should navigate back to board page
    await expect(page).toHaveURL(`/projects/${ctx.projectId}/board/${actualBoardId}`, { timeout: 10000 });

    // Should see kanban columns
    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible({ timeout: 10000 });
  });

  test('board name is consistent across tabs', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tab');
    await navigateToBoard(page, ctx.projectId);

    // Get board name from board page
    const boardName = await page.locator('h1').first().textContent();

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Navigate to Planning
    await page.getByRole('link', { name: 'Planning' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1').first()).toHaveText(boardName!, { timeout: 10000 });

    // Navigate to Metrics
    await page.getByRole('link', { name: 'Metrics' }).click();
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1').first()).toHaveText(boardName!, { timeout: 10000 });
  });

  test('active tab is highlighted correctly', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tab');
    await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // On board page, Board tab should be active (has bg-white class, no href)
    // The active tab is a span, not a link
    const boardTab = page.locator('.bg-white.shadow-sm').filter({ hasText: 'Board' });
    await expect(boardTab).toBeVisible({ timeout: 5000 });

    // Planning and Metrics should be links
    await expect(page.getByRole('link', { name: 'Planning' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Metrics' })).toBeVisible();

    // Navigate to Planning
    await page.getByRole('link', { name: 'Planning' }).click();
    await page.waitForLoadState('networkidle');

    // Planning tab should now be active
    const planningTab = page.locator('.bg-white.shadow-sm').filter({ hasText: 'Planning' });
    await expect(planningTab).toBeVisible({ timeout: 5000 });

    // Board should now be a link
    await expect(page.getByRole('link', { name: 'Board' })).toBeVisible();
  });
});

test.describe('Sprint Planning View', () => {

  test('shows backlog section by default', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Should see Backlog section (expanded by default)
    await expect(page.getByText('Backlog', { exact: true }).first()).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('No cards in backlog')).toBeVisible({ timeout: 5000 });
  });

  test('shows active sprint section when sprint exists', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create and start a sprint
    await createSprint(page, `Active Plan Sprint ${ctx.testId}`);
    await startSprint(page, `Active Plan Sprint ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Should see the active sprint section with "Active" badge
    await expect(page.getByText(`Active Plan Sprint ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Active', { exact: true }).first()).toBeVisible({ timeout: 5000 });
  });

  test('shows future sprints section', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a future sprint (don't start it)
    await createSprint(page, `Future Plan Sprint ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Should see the future sprint section with "Future" badge
    await expect(page.getByText(`Future Plan Sprint ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Future', { exact: true }).first()).toBeVisible({ timeout: 5000 });
  });

  test('backlog cards appear in backlog section', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a card (it will be in backlog since no sprints)
    await createCard(page, 'Todo', `Planning Backlog Card ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // The card should appear in the backlog section
    await expect(page.getByText(`Planning Backlog Card ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
  });

  test('can click on card to open detail panel', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a card
    await createCard(page, 'Todo', `Click Test Card ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Click on the card
    await page.getByText(`Click Test Card ${ctx.testId}`).click();

    // Card detail panel should open
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#detail-title')).toHaveValue(`Click Test Card ${ctx.testId}`);
  });

  test('URL updates when opening a card', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a card
    await createCard(page, 'Todo', `URL Card ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Click on the card
    await page.getByText(`URL Card ${ctx.testId}`).click();

    // Wait for panel to open
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // URL should have card query parameter
    await expect(page).toHaveURL(/\?card=[a-f0-9-]+/, { timeout: 5000 });
  });

  test('ESC key closes card detail panel', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a card
    await createCard(page, 'Todo', `ESC Test Card ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Click on the card
    await page.getByText(`ESC Test Card ${ctx.testId}`).click();

    // Card detail panel should open
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Press ESC
    await page.keyboard.press('Escape');

    // Panel should close
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // URL should no longer have card parameter
    await expect(page).not.toHaveURL(/\?card=/);
  });

  test('can collapse and expand sections', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Backlog should be expanded by default - shows "No cards in backlog"
    await expect(page.getByText('No cards in backlog')).toBeVisible({ timeout: 5000 });

    // Click on the Backlog header to collapse it
    await page.getByText('Backlog', { exact: true }).first().click();

    // Content should be hidden
    await expect(page.getByText('No cards in backlog')).not.toBeVisible({ timeout: 5000 });

    // Click again to expand
    await page.getByText('Backlog', { exact: true }).first().click();

    // Content should be visible again
    await expect(page.getByText('No cards in backlog')).toBeVisible({ timeout: 5000 });
  });

  test('shows card count and story points in section headers', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plan');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a card with story points
    await createCard(page, 'Todo', `Points Card ${ctx.testId}`);

    // Open the card and set story points
    await page.getByText(`Points Card ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    const storyPointsInput = page.locator('input[id$="storyPoints"]');
    await storyPointsInput.fill('5');
    await page.waitForTimeout(1500); // Wait for auto-save
    await page.getByRole('button', { name: 'Close' }).click();

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Should show card count and story points in the backlog header
    // The header shows "1 card | 5 pts" or similar
    await expect(page.getByText(/1.*card.*5.*pts/i)).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Planning Page Card Actions', () => {

  test('can move card from backlog to sprint via menu', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plact');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a sprint
    await createSprint(page, `Target Sprint ${ctx.testId}`);

    // Create a card
    await createCard(page, 'Todo', `Move Card ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Find the card row - it's a div with role="button" containing the card title
    const cardRow = page.locator('[role="button"]').filter({ hasText: `Move Card ${ctx.testId}` }).first();
    await expect(cardRow).toBeVisible({ timeout: 5000 });

    // Hover over the card row to reveal the menu button (it has opacity-0 by default)
    await cardRow.hover();

    // Click the dropdown menu trigger - it's inside [data-dropdown] div
    const menuTrigger = cardRow.locator('[data-dropdown] button');
    await expect(menuTrigger).toBeVisible({ timeout: 3000 });
    await menuTrigger.click();

    // Wait for menu to appear
    await expect(page.getByText('View Details')).toBeVisible({ timeout: 5000 });

    // Hover over "Move to Sprint" to open submenu
    await page.getByText('Move to Sprint').hover();

    // Wait for submenu to appear and click on the target sprint
    await expect(page.getByText(`Target Sprint ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
    await page.getByText(`Target Sprint ${ctx.testId}`).click();

    // Wait for success toast
    await expect(page.getByText('Card moved to sprint')).toBeVisible({ timeout: 10000 });
  });

  test('can open card details from menu', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'plact');
    const boardId = await navigateToBoard(page, ctx.projectId);

    // Get board ID from URL
    const boardUrl = page.url();
    const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
    const actualBoardId = boardMatch ? boardMatch[1] : '';

    // Create a card
    await createCard(page, 'Todo', `Details Card ${ctx.testId}`);

    // Navigate to planning
    await navigateToPlanning(page, ctx.projectId, actualBoardId);

    // Find the card row
    const cardRow = page.locator('[role="button"]').filter({ hasText: `Details Card ${ctx.testId}` }).first();
    await expect(cardRow).toBeVisible({ timeout: 5000 });

    // Hover to reveal menu
    await cardRow.hover();

    // Click the dropdown menu trigger
    const menuTrigger = cardRow.locator('[data-dropdown] button');
    await menuTrigger.click();

    // Click "View Details"
    await expect(page.getByText('View Details')).toBeVisible({ timeout: 5000 });
    await page.getByText('View Details').click();

    // Card detail panel should open
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
  });
});
