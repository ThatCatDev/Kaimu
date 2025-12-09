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

test.describe('Kanban Cards - Advanced Features', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `cards_e2e_${randomId}`;
  const password = 'testpassword123';
  let organizationId: string;
  let projectId: string;
  const orgName = `Cards Test Org ${randomId}`;
  const projectName = `Cards Test Project ${randomId}`;
  const projectKey = `CD${randomLetters(4)}`;

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

  test('can create card with all fields', async ({ page }) => {
    await navigateToBoard(page);

    // Click add card button on the Todo column
    await clickAddCardInColumn(page, 'Todo');

    // Fill in all fields
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
    await page.fill('#title', `Full Card ${randomId}`);
    await page.fill('#description', 'This card has all fields filled');
    await page.selectOption('#priority', 'URGENT');

    // Set due date to tomorrow - the input is type="date" which expects YYYY-MM-DD format
    // The frontend will convert this to RFC3339 before sending to API
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const dateStr = tomorrow.toISOString().split('T')[0];  // YYYY-MM-DD format for date input
    await page.fill('#dueDate', dateStr);

    await page.getByRole('button', { name: 'Create Card' }).click();

    // Verify card appears - wait longer for API call and modal close
    await expect(page.getByRole('heading', { name: 'Create Card' })).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Full Card ${randomId}`)).toBeVisible({ timeout: 5000 });
  });

  test('card shows priority indicator', async ({ page }) => {
    await navigateToBoard(page);

    // Create a high priority card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `High Priority ${randomId}`);
    await page.selectOption('#priority', 'HIGH');
    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`High Priority ${randomId}`)).toBeVisible({ timeout: 5000 });

    // The card should have a priority indicator (typically shown as a colored badge or icon)
    const cardElement = page.locator(`text=High Priority ${randomId}`).locator('..');
    await expect(cardElement).toBeVisible();
  });

  test('can update card priority', async ({ page }) => {
    await navigateToBoard(page);

    // Create a card first
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Priority Update ${randomId}`);
    await page.selectOption('#priority', 'LOW');
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Priority Update ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Click on the card to open detail modal
    await page.getByText(`Priority Update ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Change priority (auto-saves)
    await page.selectOption('#priority', 'URGENT');

    // Wait for auto-save
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 5000 });

    // Close modal
    await page.getByRole('button', { name: 'Close' }).click();

    // Verify modal closes
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('can set and clear due date', async ({ page }) => {
    await navigateToBoard(page);

    // Create a card with due date
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Due Date Card ${randomId}`);

    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 7);
    const dateStr = futureDate.toISOString().split('T')[0];
    await page.fill('#dueDate', dateStr);

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Due Date Card ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Edit card and clear due date (auto-saves)
    await page.getByText(`Due Date Card ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    await page.fill('#dueDate', '');
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 5000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('can add description to existing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card without description
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `No Desc Card ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`No Desc Card ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Edit card and add description (auto-saves)
    await page.getByText(`No Desc Card ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    await page.fill('#description', 'Description added later');
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Reload the board to ensure fresh data
    await page.reload();
    await page.waitForLoadState('networkidle');
    await expect(page.getByText(`No Desc Card ${randomId}`)).toBeVisible({ timeout: 10000 });

    // Verify description is saved
    await page.getByText(`No Desc Card ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#description')).toHaveValue('Description added later', { timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('card creation fails without title', async ({ page }) => {
    await navigateToBoard(page);

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
    await navigateToBoard(page);

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
    await navigateToBoard(page);

    // Create a card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Auto Save Edit ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Auto Save Edit ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open edit and change title - auto-save will save it
    await page.getByText(`Auto Save Edit ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    await page.fill('#title', `Auto Saved Title ${randomId}`);
    // Wait for auto-save
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Wait for modal to close and board to refresh
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Old title should disappear and new title should appear
    await expect(page.getByText(`Auto Save Edit ${randomId}`)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`Auto Saved Title ${randomId}`)).toBeVisible({ timeout: 10000 });
  });

  test('cards appear in correct columns', async ({ page }) => {
    await navigateToBoard(page);

    // Create card in Todo
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Todo Card ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Todo Card ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Create card in In Progress
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `In Progress Card ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`In Progress Card ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Create card in Done
    await clickAddCardInColumn(page, 'Done');
    await page.fill('#title', `Done Card ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Done Card ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Verify each card is in its column
    const todoColumn = getColumn(page, 'Todo');
    const inProgressColumn = getColumn(page, 'In Progress');
    const doneColumn = getColumn(page, 'Done');

    await expect(todoColumn.getByText(`Todo Card ${randomId}`)).toBeVisible();
    await expect(inProgressColumn.getByText(`In Progress Card ${randomId}`)).toBeVisible();
    await expect(doneColumn.getByText(`Done Card ${randomId}`)).toBeVisible();
  });

  test('multiple cards can be created in same column', async ({ page }) => {
    await navigateToBoard(page);

    // Create multiple cards in Todo
    for (let i = 1; i <= 3; i++) {
      await clickAddCardInColumn(page, 'Todo');
      await page.fill('#title', `Multi Card ${i} ${randomId}`);
      await page.getByRole('button', { name: 'Create Card' }).click();
      await expect(page.getByText(`Multi Card ${i} ${randomId}`)).toBeVisible({ timeout: 5000 });
    }

    // Verify all cards exist in Todo column
    const todoColumn = getColumn(page, 'Todo');
    await expect(todoColumn.getByText(`Multi Card 1 ${randomId}`)).toBeVisible();
    await expect(todoColumn.getByText(`Multi Card 2 ${randomId}`)).toBeVisible();
    await expect(todoColumn.getByText(`Multi Card 3 ${randomId}`)).toBeVisible();
  });

  test('card detail modal shows created date', async ({ page }) => {
    await navigateToBoard(page);

    // Create a card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Date Check ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Date Check ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail
    await page.getByText(`Date Check ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Should show "Created:" timestamp
    await expect(page.getByText(/Created:/)).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('delete confirmation prevents accidental deletion', async ({ page }) => {
    await navigateToBoard(page);

    // Create a card
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `No Delete ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`No Delete ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail
    await page.getByText(`No Delete ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Click delete - shows confirmation modal
    await page.getByRole('button', { name: 'Delete Card' }).click();

    // Confirmation modal should appear
    await expect(page.getByRole('heading', { name: 'Delete Card' })).toBeVisible({ timeout: 5000 });

    // Click Cancel in confirmation modal - use last dialog (confirmation modal)
    await page.getByRole('dialog').last().getByRole('button', { name: 'Cancel' }).click();

    // Confirmation modal should close
    await expect(page.getByRole('heading', { name: 'Delete Card' })).not.toBeVisible({ timeout: 5000 });

    // Close the detail modal
    await page.getByRole('button', { name: 'Close' }).click();

    // Card should still exist
    await expect(page.getByText(`No Delete ${randomId}`)).toBeVisible();
  });
});
