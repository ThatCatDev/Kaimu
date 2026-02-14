import { test, expect } from '@playwright/test';
import { setupTestEnvironment, navigateToBoard, createCard, getColumn } from './helpers';

test.describe('Bulk Card Actions', () => {

  test('selection mode toggle shows and hides', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'bulk');
    await navigateToBoard(page, ctx.projectId);

    // Click "Select" to enter selection mode
    const selectButton = page.getByRole('button', { name: 'Select', exact: true });
    await expect(selectButton).toBeVisible({ timeout: 5000 });
    await selectButton.click();

    // Button text should change to "Cancel"
    await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible({ timeout: 5000 });

    // Click "Cancel" to exit selection mode
    await page.getByRole('button', { name: 'Cancel', exact: true }).click();

    // Button text should revert to "Select"
    await expect(page.getByRole('button', { name: 'Select', exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('can select cards and see toolbar', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'bulk');
    await navigateToBoard(page, ctx.projectId);

    // Create 2 cards in Todo
    await createCard(page, 'Todo', `Select Card A ${ctx.testId}`);
    await createCard(page, 'Todo', `Select Card B ${ctx.testId}`);

    // Enter selection mode
    await page.getByRole('button', { name: 'Select', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible({ timeout: 5000 });

    // Wait for checkboxes to appear
    await expect(page.locator('input[type="checkbox"]').first()).toBeVisible({ timeout: 5000 });

    // Click on each card to select it (clicking the card toggles selection in selection mode)
    await page.getByText(`Select Card A ${ctx.testId}`).click();
    await page.getByText(`Select Card B ${ctx.testId}`).click();

    // Verify toolbar shows "2 selected"
    await expect(page.getByText('2 selected')).toBeVisible({ timeout: 10000 });

    // Click "Clear" to deselect all
    await page.getByText('Clear').click();

    // Toolbar should disappear (selectedCount becomes 0)
    await expect(page.getByText('2 selected')).not.toBeVisible({ timeout: 5000 });
  });

  test('bulk move to column', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'bulk');
    await navigateToBoard(page, ctx.projectId);

    // Create 2 cards in Todo
    await createCard(page, 'Todo', `Move Card A ${ctx.testId}`);
    await createCard(page, 'Todo', `Move Card B ${ctx.testId}`);

    // Verify cards are in Todo column
    const todoColumn = getColumn(page, 'Todo');
    await expect(todoColumn.getByText(`Move Card A ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
    await expect(todoColumn.getByText(`Move Card B ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Enter selection mode
    await page.getByRole('button', { name: 'Select', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible({ timeout: 5000 });

    // Select both cards
    await page.getByText(`Move Card A ${ctx.testId}`).click();
    await page.getByText(`Move Card B ${ctx.testId}`).click();
    await expect(page.getByText('2 selected')).toBeVisible({ timeout: 10000 });

    // Click "Column" dropdown in toolbar
    await page.getByRole('button', { name: 'Column', exact: true }).click();

    // Click "In Progress" option from the dropdown menu
    await page.getByRole('menuitem', { name: 'In Progress' }).click();

    // Verify toast message
    await expect(page.getByText('Moved 2 card(s)')).toBeVisible({ timeout: 10000 });

    // Wait for board to refresh
    await page.waitForTimeout(1000);

    // Verify cards now appear in In Progress column
    const inProgressColumn = getColumn(page, 'In Progress');
    await expect(inProgressColumn.getByText(`Move Card A ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
    await expect(inProgressColumn.getByText(`Move Card B ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
  });

  test('bulk set priority', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'bulk');
    await navigateToBoard(page, ctx.projectId);

    // Create 2 cards in Todo
    await createCard(page, 'Todo', `Priority Card A ${ctx.testId}`);
    await createCard(page, 'Todo', `Priority Card B ${ctx.testId}`);

    // Enter selection mode
    await page.getByRole('button', { name: 'Select', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible({ timeout: 5000 });

    // Select both cards
    await page.getByText(`Priority Card A ${ctx.testId}`).click();
    await page.getByText(`Priority Card B ${ctx.testId}`).click();
    await expect(page.getByText('2 selected')).toBeVisible({ timeout: 10000 });

    // Click "Priority" dropdown in toolbar
    await page.getByRole('button', { name: 'Priority', exact: true }).click();

    // Click "High" option
    await page.getByText('High').click();

    // Verify toast message
    await expect(page.getByText('Updated priority for 2 card(s)')).toBeVisible({ timeout: 10000 });
  });

  test('bulk set assignee', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'bulk');
    await navigateToBoard(page, ctx.projectId);

    // Create 2 cards in Todo
    await createCard(page, 'Todo', `Assignee Card A ${ctx.testId}`);
    await createCard(page, 'Todo', `Assignee Card B ${ctx.testId}`);

    // Enter selection mode
    await page.getByRole('button', { name: 'Select', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible({ timeout: 5000 });

    // Select both cards
    await page.getByText(`Assignee Card A ${ctx.testId}`).click();
    await page.getByText(`Assignee Card B ${ctx.testId}`).click();
    await expect(page.getByText('2 selected')).toBeVisible({ timeout: 10000 });

    // Click "Assignee" dropdown in toolbar
    await page.getByRole('button', { name: 'Assignee', exact: true }).click();

    // Verify "Unassign" option is visible
    await expect(page.getByText('Unassign')).toBeVisible({ timeout: 5000 });

    // Verify at least one member name is shown (the test user who created the org)
    // The dropdown should have more than just the "Unassign" option
    const dropdownItems = page.locator('[data-dropdown-menu-content] [data-dropdown-menu-item]');
    await expect(dropdownItems.first()).toBeVisible({ timeout: 5000 });
    const itemCount = await dropdownItems.count();
    expect(itemCount).toBeGreaterThanOrEqual(2); // "Unassign" + at least one member

    // Click the last member (not Unassign) - it should be after the separator
    await dropdownItems.last().click();

    // Verify toast message
    await expect(page.getByText('Updated assignee for 2 card(s)')).toBeVisible({ timeout: 10000 });
  });

  test('bulk delete with confirmation', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'bulk');
    await navigateToBoard(page, ctx.projectId);

    // Create 3 cards in Todo
    await createCard(page, 'Todo', `Delete Card A ${ctx.testId}`);
    await createCard(page, 'Todo', `Delete Card B ${ctx.testId}`);
    await createCard(page, 'Todo', `Keep Card C ${ctx.testId}`);

    // Enter selection mode
    await page.getByRole('button', { name: 'Select', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeVisible({ timeout: 5000 });

    // Select only the first 2 cards (not the 3rd)
    await page.getByText(`Delete Card A ${ctx.testId}`).click();
    await page.getByText(`Delete Card B ${ctx.testId}`).click();
    await expect(page.getByText('2 selected')).toBeVisible({ timeout: 10000 });

    // Click "Delete" button in toolbar
    await page.getByRole('button', { name: 'Delete', exact: true }).click();

    // Verify confirmation modal appears with "Delete 2 Cards"
    await expect(page.getByRole('heading', { name: 'Delete 2 Cards' })).toBeVisible({ timeout: 5000 });

    // Click "Delete" in confirmation modal
    await page.getByRole('button', { name: 'Delete', exact: true }).last().click();

    // Verify toast message
    await expect(page.getByText('Deleted 2 card(s)')).toBeVisible({ timeout: 10000 });

    // Wait for board to refresh
    await page.waitForTimeout(1000);

    // Verify the 2 deleted cards are gone
    await expect(page.getByText(`Delete Card A ${ctx.testId}`)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Delete Card B ${ctx.testId}`)).not.toBeVisible({ timeout: 10000 });

    // Verify the 3rd card still exists
    await expect(page.getByText(`Keep Card C ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
  });

});
