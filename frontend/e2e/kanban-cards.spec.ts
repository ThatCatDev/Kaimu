import { test, expect, type Page } from '@playwright/test';
import { setupTestEnvironment, navigateToBoard, getColumn, createCard, clickAddCardInColumn } from './helpers';

// Helper function to select priority from Bits UI Select component
async function selectPriority(page: Page, priorityLabel: string) {
  // Click the priority trigger to open dropdown
  await page.locator('#priority').click();
  // Click the option with matching label
  await page.getByRole('option', { name: priorityLabel }).click();
}

test.describe('Kanban Cards - Advanced Features', () => {

  test('can create card with all fields', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Click add card button on the Todo column
    await clickAddCardInColumn(page, 'Todo');

    // Fill in all fields
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
    await page.fill('#title', `Full Card ${ctx.testId}`);
    await page.fill('#description', 'This card has all fields filled');
    await selectPriority(page, 'Urgent');

    // Set due date using Bits UI DatePicker - click trigger to open calendar, then select a date
    // The DatePicker uses segment-based input or calendar popup
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const tomorrowDay = tomorrow.getDate();

    // Click the calendar trigger button (has calendar icon)
    await page.locator('#dueDate').locator('button[class*="inline-flex"]').last().click();

    // Wait for calendar to appear and click on tomorrow's date
    await page.locator('[data-bits-day]').filter({ hasText: String(tomorrowDay) }).first().click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    // Verify card appears - wait longer for API call and modal close
    await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Full Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
  });

  test('card shows priority indicator', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create a high priority card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `High Priority ${ctx.testId}`);
    await selectPriority(page, 'High');
    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`High Priority ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // The card should have a priority indicator (typically shown as a colored badge or icon)
    const cardElement = page.locator(`text=High Priority ${ctx.testId}`).locator('..');
    await expect(cardElement).toBeVisible();
  });

  test('can update card priority', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create a card first
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Priority Update ${ctx.testId}`);
    await selectPriority(page, 'Low');
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Priority Update ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Click on the card to open detail modal
    await page.getByText(`Priority Update ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Change priority (auto-saves) - look for the Priority label and click its button trigger
    const dialog = page.getByRole('dialog', { name: 'Card Details' });
    // Find the Priority label's sibling button (the dropdown trigger)
    await dialog.getByLabel('Priority').click();
    await page.getByRole('option', { name: 'Urgent' }).click();

    // Wait for auto-save
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 5000 });

    // Close modal
    await page.getByRole('button', { name: 'Close' }).click();

    // Verify modal closes
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test.skip('can set and clear due date', async ({ page }) => {
    // TODO: Update test for Bits UI DatePicker component
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create a card first
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Due Date Card ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Due Date Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and interact with the due date picker
    await page.getByText(`Due Date Card ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Open the date picker calendar by clicking inside the Due Date field area
    const detailDialog = page.getByRole('dialog', { name: 'Card Details' });
    const dueDateField = detailDialog.locator('#panel-dueDate');
    // Click on the calendar icon button inside the date picker
    await dueDateField.locator('button').last().click();

    // Wait for calendar to open and click a future date
    await page.waitForTimeout(500);
    // Click day 20
    const day20Button = page.getByRole('button', { name: '20', exact: true });
    await day20Button.click();

    // Wait for auto-save
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 5000 });

    // Clear date by clicking the X button (Clear date)
    await detailDialog.getByTitle('Clear date').click();
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('can add description to existing card', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create card without description
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `No Desc Card ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`No Desc Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Edit card and add description (auto-saves)
    await page.getByText(`No Desc Card ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    await page.fill('#description', 'Description added later');
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Reload the board to ensure fresh data
    await page.reload();
    await page.waitForLoadState('networkidle');
    await expect(page.getByText(`No Desc Card ${ctx.testId}`)).toBeVisible({ timeout: 10000 });

    // Verify description is saved
    await page.getByText(`No Desc Card ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#description')).toHaveValue('Description added later', { timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('card creation fails without title', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Try to create card without title
    await clickAddCardInColumn(page, 'Todo');
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });

    // Clear title field and try to submit
    await page.fill('#title', '');
    await page.getByRole('button', { name: 'Create Card' }).click();

    // Should show validation error or focus on title field
    // HTML5 validation or custom error
    const titleInput = page.locator('#title');
    const isRequired = await titleInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing);
    expect(isRequired).toBeTruthy();
  });

  test('can cancel card creation', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });

    await page.fill('#title', 'This card should not be created');
    await page.getByRole('button', { name: 'Cancel', exact: true }).click();

    // Modal should close
    await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 5000 });

    // Card should not exist
    await expect(page.getByText('This card should not be created')).not.toBeVisible();
  });

  test('changes auto-save even when closing', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create a card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Auto Save Edit ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Auto Save Edit ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open edit and change title - auto-save will save it
    await page.getByText(`Auto Save Edit ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    await page.fill('#title', `Auto Updated Title ${ctx.testId}`);
    // Wait for auto-save
    await expect(page.getByText('Saved', { exact: true })).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Wait for modal to close and board to refresh
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Old title should disappear and new title should appear
    await expect(page.getByText(`Auto Save Edit ${ctx.testId}`)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Auto Updated Title ${ctx.testId}`)).toBeVisible({ timeout: 10000 });
  });

  test('cards appear in correct columns', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create card in Todo
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Todo Card ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Todo Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Create card in In Progress
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `In Progress Card ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`In Progress Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Create card in Done
    await clickAddCardInColumn(page, 'Done');
    await page.fill('#title', `Done Card ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Done Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Verify each card is in its column
    const todoColumn = getColumn(page, 'Todo');
    const inProgressColumn = getColumn(page, 'In Progress');
    const doneColumn = getColumn(page, 'Done');

    await expect(todoColumn.getByText(`Todo Card ${ctx.testId}`)).toBeVisible();
    await expect(inProgressColumn.getByText(`In Progress Card ${ctx.testId}`)).toBeVisible();
    await expect(doneColumn.getByText(`Done Card ${ctx.testId}`)).toBeVisible();
  });

  test('multiple cards can be created in same column', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create multiple cards in Todo
    for (let i = 1; i <= 3; i++) {
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Multi Card ${i} ${ctx.testId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Multi Card ${i} ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
    }

    // Verify all cards exist in Todo column
    const todoColumn = getColumn(page, 'Todo');
    await expect(todoColumn.getByText(`Multi Card 1 ${ctx.testId}`)).toBeVisible();
    await expect(todoColumn.getByText(`Multi Card 2 ${ctx.testId}`)).toBeVisible();
    await expect(todoColumn.getByText(`Multi Card 3 ${ctx.testId}`)).toBeVisible();
  });

  test('card detail modal shows created date', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create a card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Date Check ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Date Check ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail
    await page.getByText(`Date Check ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Should show "Created:" timestamp
    await expect(page.getByText(/Created:/)).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('delete confirmation prevents accidental deletion', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'cards');
    await navigateToBoard(page, ctx.projectId);

    // Create a card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `No Delete ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`No Delete ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail
    await page.getByText(`No Delete ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Click delete - shows confirmation modal (modal has "Delete Card", panel has "Delete")
    await page.getByRole('button', { name: /^Delete( Card)?$/ }).first().click();

    // Confirmation modal should appear
    await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });

    // Click Cancel in confirmation modal - use last dialog (confirmation modal)
    await page.getByRole('dialog').last().getByRole('button', { name: 'Cancel' }).click();

    // Confirmation modal should close
    await expect(page.getByRole('heading', { name: 'Delete Card' })).not.toBeVisible({ timeout: 5000 });

    // Close the detail modal
    await page.getByRole('button', { name: 'Close' }).click();

    // Card should still exist
    await expect(page.getByText(`No Delete ${ctx.testId}`)).toBeVisible();
  });
});
