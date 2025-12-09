import { test, expect } from '@playwright/test';
import { setupTestEnvironment, navigateToBoard, getColumn, createCard, type TestContext } from './helpers';

test.describe('Kanban Board', () => {
  test('project detail page shows kanban board link', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await page.goto(`/projects/${ctx.projectId}`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('main').getByRole('heading', { name: 'Kanban Board' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('main').getByText('Default Board')).toBeVisible();
  });

  test('can navigate to kanban board from project', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await page.goto(`/projects/${ctx.projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();

    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'Default Board' })).toBeVisible({ timeout: 10000 });
  });

  test('kanban board shows default columns', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await navigateToBoard(page, ctx.projectId);

    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'In Progress' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Done' })).toBeVisible();
  });

  test('can toggle hidden columns visibility', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await navigateToBoard(page, ctx.projectId);

    // Backlog should be hidden by default
    await expect(page.getByRole('heading', { name: 'Backlog' })).not.toBeVisible();

    // Toggle show hidden columns
    await page.getByLabel('Show hidden columns').click();

    // Backlog should now be visible
    await expect(page.getByRole('heading', { name: 'Backlog' })).toBeVisible({ timeout: 5000 });
  });

  test('can create a new card', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await navigateToBoard(page, ctx.projectId);

    // Click add card button on the Todo column
    const todoColumn = getColumn(page, 'Todo');
    await todoColumn.getByRole('button', { name: 'Add card' }).click();

    // Fill in the card form
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
    await page.fill('#title', `Test Card ${ctx.testId}`);
    await page.fill('#description', 'This is a test card');
    await page.locator('#priority').click();
    await page.getByRole('option', { name: 'High' }).click();
    await page.getByRole('button', { name: 'Create Card' }).click();

    // Modal should close and card should appear
    await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 5000 });
    await expect(page.getByText(`Test Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
  });

  test('can view and edit card details', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await navigateToBoard(page, ctx.projectId);

    // Create a card first
    await createCard(page, 'Todo', `View Card ${ctx.testId}`);

    // Click on the card
    await page.getByText(`View Card ${ctx.testId}`).click();

    // Card detail modal should open
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#title')).toHaveValue(`View Card ${ctx.testId}`);

    // Update the card (auto-saves)
    await page.fill('#title', `Updated Card ${ctx.testId}`);
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Modal should close and updated card should appear
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
    await expect(page.getByText(`View Card ${ctx.testId}`)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Updated Card ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
  });

  test('can delete a card', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await navigateToBoard(page, ctx.projectId);

    // Create a card
    await createCard(page, 'Todo', `Delete Me ${ctx.testId}`);

    // Click on the card to open detail modal
    await page.getByText(`Delete Me ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Click delete button
    await page.getByRole('button', { name: /^Delete( Card)?$/ }).first().click();

    // Confirmation modal appears
    await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });

    // Confirm deletion
    await page.getByRole('dialog').last().getByRole('button', { name: 'Delete', exact: true }).click();

    // Card should be gone
    await expect(page.getByText(`Delete Me ${ctx.testId}`)).not.toBeVisible({ timeout: 5000 });
  });

  test('can navigate back to project via sidebar', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'board');

    await navigateToBoard(page, ctx.projectId);

    // Click on project in sidebar to navigate back
    const sidebar = page.locator('aside');
    await sidebar.getByRole('link', { name: ctx.projectName }).click();
    await expect(page).toHaveURL(`/projects/${ctx.projectId}`);
  });
});
