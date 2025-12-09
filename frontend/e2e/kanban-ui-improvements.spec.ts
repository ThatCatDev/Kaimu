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

test.describe('Kanban UI Improvements', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `ui_e2e_${randomId}`;
  const password = 'testpassword123';
  let organizationId: string;
  let projectId: string;
  const orgName = `UI Test Org ${randomId}`;
  const projectName = `UI Test Project ${randomId}`;
  const projectKey = `UI${randomLetters(4)}`;

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

  // Helper function to navigate to the board
  async function navigateToBoard(page: any) {
    await page.goto(`/projects/${projectId}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: /Kanban Board/ }).click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'Todo', exact: true })).toBeVisible({ timeout: 10000 });
  }

  // Helper function to get a column by name
  function getColumn(page: any, columnName: string) {
    return page.locator('.w-72').filter({ has: page.locator(`h3:has-text("${columnName}")`) });
  }

  // Helper to click add card button in column
  async function clickAddCardInColumn(page: any, columnName: string) {
    await getColumn(page, columnName).getByRole('button', { name: 'Add card' }).click();
  }

  test.describe('Add Card Button Location', () => {
    test('add card button is inside column, not in header', async ({ page }) => {
      await navigateToBoard(page);

      const todoColumn = getColumn(page, 'Todo');

      // Add card button should be in the column body, not header
      const addButton = todoColumn.getByRole('button', { name: 'Add card' });
      await expect(addButton).toBeVisible();

      // The button should have text "Add card" (not just an icon)
      await expect(addButton).toContainText('Add card');
    });

    test('add card button works in each column', async ({ page }) => {
      await navigateToBoard(page);

      // Test Todo column
      await clickAddCardInColumn(page, 'Todo');
      await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
      await page.getByRole('button', { name: 'Cancel' }).click();

      // Test In Progress column
      await clickAddCardInColumn(page, 'In Progress');
      await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
      await page.getByRole('button', { name: 'Cancel' }).click();

      // Test Done column
      await clickAddCardInColumn(page, 'Done');
      await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
      await page.getByRole('button', { name: 'Cancel' }).click();
    });
  });

  test.describe('Auto-save Functionality', () => {
    test('card auto-saves when editing', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card first
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Auto Save Test ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Auto Save Test ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Auto Save Test ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Change the title
      await page.fill('#title', `Auto Save Updated ${randomId}`);

      // Should show "Saved" indicator after auto-save completes
      await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });

      // Close the modal - this triggers onUpdated which refreshes the board
      await page.getByRole('button', { name: 'Close' }).click();

      // Wait for modal to close and board to refresh
      await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

      // Old title should disappear and new title should appear
      await expect(page.getByText(`Auto Save Test ${randomId}`)).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(`Auto Save Updated ${randomId}`)).toBeVisible({ timeout: 10000 });
    });

    test('shows saving indicator while saving', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Saving Indicator ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Saving Indicator ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Saving Indicator ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Make a change and check for saving/saved indicator
      await page.fill('#description', 'Testing auto save indicator');

      // Should eventually show "Saved"
      await expect(page.getByText('Saved')).toBeVisible({ timeout: 5000 });

      await page.getByRole('button', { name: 'Close' }).click();
    });

    test('footer shows auto-save hint', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Footer Hint ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Footer Hint ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Footer Hint ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Should show auto-save hint in footer
      await expect(page.getByText('Auto-saves as you type')).toBeVisible();

      await page.getByRole('button', { name: 'Close' }).click();
    });
  });

  test.describe('Keyboard Shortcuts', () => {
    test('Escape key closes card detail modal', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Escape Test ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Escape Test ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Escape Test ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Press Escape
      await page.keyboard.press('Escape');

      // Modal should close
      await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
    });

    test('Escape key closes create card modal', async ({ page }) => {
      await navigateToBoard(page);

      await clickAddCardInColumn(page, 'Todo');
      await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });

      // Press Escape
      await page.keyboard.press('Escape');

      // Modal should close
      await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 5000 });
    });

    test('modal shows Escape hint', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Hint Test ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Hint Test ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Hint Test ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Should show Escape hint - use exact match and look in kbd element
      await expect(page.locator('kbd').filter({ hasText: 'Esc' })).toBeVisible();
      await expect(page.getByText('to close')).toBeVisible();

      await page.keyboard.press('Escape');
    });
  });

  test.describe('Delete Confirmation Modal', () => {
    test('delete button shows confirmation modal instead of browser dialog', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Delete Modal Test ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Delete Modal Test ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Delete Modal Test ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Click delete
      await page.getByRole('button', { name: 'Delete Card' }).click();

      // Should show confirmation modal (not browser dialog)
      await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });
      await expect(page.getByText('Are you sure you want to delete this card?')).toBeVisible();
      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
      await expect(page.getByRole('button', { name: 'Delete', exact: true })).toBeVisible();
    });

    test('cancel in delete confirmation keeps card', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Keep Card ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Keep Card ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Keep Card ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Click delete, then cancel
      await page.getByRole('button', { name: 'Delete Card' }).click();
      await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });
      await page.getByRole('dialog').last().getByRole('button', { name: 'Cancel' }).click();

      // Confirmation modal should close
      await expect(page.getByRole('heading', { name: 'Delete Card' })).not.toBeVisible({ timeout: 5000 });

      // Card detail should still be open
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible();

      // Close and verify card still exists
      await page.getByRole('button', { name: 'Close' }).click();
      await expect(page.getByText(`Keep Card ${randomId}`)).toBeVisible();
    });

    test('confirm delete removes card', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Delete Me ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Delete Me ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Delete Me ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Click delete, then confirm
      await page.getByRole('button', { name: 'Delete Card' }).click();
      await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });

      // Click the Delete button in the confirmation modal
      await page.getByRole('dialog').last().getByRole('button', { name: 'Delete', exact: true }).click();

      // Modal should close and card should be gone
      await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
      await expect(page.getByText(`Delete Me ${randomId}`)).not.toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Quick Actions on Card Hover', () => {
    test('card shows quick action buttons on hover', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Hover Actions ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Hover Actions ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Find the card and hover over it
      const card = page.locator(`text=Hover Actions ${randomId}`).locator('..');
      await card.hover();

      // Should show edit and delete buttons
      await expect(card.locator('button[title="Edit card"]')).toBeVisible({ timeout: 2000 });
      await expect(card.locator('button[title="Delete card"]')).toBeVisible({ timeout: 2000 });
    });

    test('quick edit opens card detail', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Quick Edit ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Quick Edit ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Hover and click edit
      const card = page.locator(`text=Quick Edit ${randomId}`).locator('..');
      await card.hover();
      await card.locator('button[title="Edit card"]').click();

      // Card detail should open
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      await page.keyboard.press('Escape');
    });

    test('quick delete shows confirmation modal', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Quick Delete ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Quick Delete ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Hover and click delete
      const card = page.locator(`text=Quick Delete ${randomId}`).locator('..');
      await card.hover();
      await card.locator('button[title="Delete card"]').click();

      // Confirmation modal should appear
      await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });
      await expect(page.getByText('Are you sure you want to delete this card?')).toBeVisible();

      // Cancel to keep the card
      await page.getByRole('dialog').last().getByRole('button', { name: 'Cancel' }).click();
      await expect(page.getByText(`Quick Delete ${randomId}`)).toBeVisible();
    });
  });

  test.describe('Side Panel View', () => {
    test('can switch from modal to panel view', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Panel View ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Panel View ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail (in modal mode by default)
      await page.getByText(`Panel View ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Click the panel view toggle button
      await page.locator('button[title="Switch to side panel view"]').click();

      // Should now be in panel mode (panel slides in from right)
      // The panel has translate-x-0 class when open
      await expect(page.locator('.fixed.inset-y-0.right-0.translate-x-0')).toBeVisible({ timeout: 5000 });

      await page.keyboard.press('Escape');
    });

    test('panel view allows selecting other cards', async ({ page }) => {
      await navigateToBoard(page);

      // Create two cards
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Panel Card 1 ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Panel Card 1 ${randomId}`)).toBeVisible({ timeout: 5000 });

      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Panel Card 2 ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Panel Card 2 ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open first card and switch to panel view
      await page.getByText(`Panel Card 1 ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
      await page.locator('button[title="Switch to side panel view"]').click();

      // Verify we're in panel view with first card
      await expect(page.locator('#panel-title')).toHaveValue(`Panel Card 1 ${randomId}`);

      // Click on second card (panel should stay open, no backdrop blocking)
      await page.getByText(`Panel Card 2 ${randomId}`).click();

      // Panel should now show second card
      await expect(page.locator('#panel-title')).toHaveValue(`Panel Card 2 ${randomId}`, { timeout: 5000 });

      await page.keyboard.press('Escape');
    });

    test('can switch from panel back to modal view', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Switch Back ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Switch Back ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open in modal and switch to panel
      await page.getByText(`Switch Back ${randomId}`).click();
      await page.locator('button[title="Switch to side panel view"]').click();
      await expect(page.locator('.fixed.inset-y-0.right-0.translate-x-0')).toBeVisible({ timeout: 5000 });

      // Switch back to modal
      await page.locator('button[title="Switch to modal view"]').click();

      // Should be back in modal mode (centered modal with backdrop)
      await expect(page.locator('.fixed.inset-0.bg-gray-900\\/60')).toBeVisible({ timeout: 5000 });

      await page.keyboard.press('Escape');
    });

    test('view mode preference is persisted', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Persist Mode ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Persist Mode ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open and switch to panel view
      await page.getByText(`Persist Mode ${randomId}`).click();
      await page.locator('button[title="Switch to side panel view"]').click();
      await expect(page.locator('.fixed.inset-y-0.right-0.translate-x-0')).toBeVisible({ timeout: 5000 });
      await page.keyboard.press('Escape');

      // Reload the page
      await page.reload();
      await expect(page.getByRole('heading', { name: 'Todo', exact: true })).toBeVisible({ timeout: 10000 });

      // Open the card again - should open in panel mode
      await page.getByText(`Persist Mode ${randomId}`).click();
      await expect(page.locator('.fixed.inset-y-0.right-0.translate-x-0')).toBeVisible({ timeout: 5000 });

      await page.keyboard.press('Escape');
    });
  });

  test.describe('Inline Label Creation', () => {
    test('can create label by typing and pressing enter', async ({ page }) => {
      await navigateToBoard(page);

      // Create a card and open it
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Label Create ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Label Create ${randomId}`)).toBeVisible({ timeout: 5000 });

      // Open card detail
      await page.getByText(`Label Create ${randomId}`).click();
      await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

      // Type in the label input
      const labelInput = page.locator('input[placeholder*="search or create labels"]');
      await labelInput.fill(`NewLabel${randomId}`);

      // Should show "Create" option in dropdown
      await expect(page.getByText(`Create "NewLabel${randomId}"`)).toBeVisible({ timeout: 5000 });

      // Press Enter to create
      await labelInput.press('Enter');

      // Label should be created and selected (shown in selected labels area)
      await expect(page.locator('.rounded-full').filter({ hasText: `NewLabel${randomId}` })).toBeVisible({ timeout: 5000 });

      await page.keyboard.press('Escape');
    });

    test('can search and select existing labels', async ({ page }) => {
      await navigateToBoard(page);

      // First create a label via a card
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `First Label Card ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await page.getByText(`First Label Card ${randomId}`).click();

      const labelInput = page.locator('input[placeholder*="search or create labels"]');
      await labelInput.fill(`SearchLabel${randomId}`);
      await page.getByText(`Create "SearchLabel${randomId}"`).click();
      await page.keyboard.press('Escape');

      // Now create another card and search for that label
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Second Label Card ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await page.getByText(`Second Label Card ${randomId}`).click();

      // Search for the existing label
      const labelInput2 = page.locator('input[placeholder*="search or create labels"]');
      await labelInput2.fill(`SearchLabel${randomId}`);

      // Should show the existing label in dropdown
      await expect(page.locator('.absolute.z-10').getByText(`SearchLabel${randomId}`)).toBeVisible({ timeout: 5000 });

      // Click to select
      await page.locator('.absolute.z-10').getByText(`SearchLabel${randomId}`).click();

      // Label should be selected
      await expect(page.locator('.rounded-full').filter({ hasText: `SearchLabel${randomId}` })).toBeVisible();

      await page.keyboard.press('Escape');
    });

    test('can remove selected label by clicking X', async ({ page }) => {
      await navigateToBoard(page);

      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Remove Label ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await page.getByText(`Remove Label ${randomId}`).click();

      // Create and select a label
      const labelInput = page.locator('input[placeholder*="search or create labels"]');
      await labelInput.fill(`RemoveMe${randomId}`);
      await labelInput.press('Enter');

      // Verify label is selected
      const selectedLabel = page.locator('.rounded-full').filter({ hasText: `RemoveMe${randomId}` });
      await expect(selectedLabel).toBeVisible({ timeout: 5000 });

      // Click the X button on the label to remove it
      await selectedLabel.locator('button').click();

      // Label should be removed from selection
      await expect(selectedLabel).not.toBeVisible({ timeout: 5000 });

      await page.keyboard.press('Escape');
    });
  });
});
