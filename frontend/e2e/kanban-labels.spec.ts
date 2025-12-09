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

test.describe('Kanban Cards with Labels', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `labels_e2e_${randomId}`;
  const password = 'testpassword123';
  let organizationId: string;
  let projectId: string;
  const orgName = `Labels Test Org ${randomId}`;
  const projectName = `Labels Test Project ${randomId}`;
  const projectKey = `LB${randomLetters(4)}`;

  test.beforeAll(async ({ browser }) => {
    // Register a user, create an organization, project, and labels
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

    // Create labels via GraphQL API using page's request context (has auth cookies)
    const labels = [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
      { name: 'Documentation', color: '#8B5CF6', description: 'Documentation needs update' },
    ];

    for (const label of labels) {
      const response = await page.request.post('http://localhost:3000/graphql', {
        data: {
          query: `
            mutation CreateLabel($input: CreateLabelInput!) {
              createLabel(input: $input) {
                id
                name
                color
              }
            }
          `,
          variables: {
            input: {
              projectId,
              name: label.name,
              color: label.color,
              description: label.description,
            },
          },
        },
      });
      // Handle response (log errors for debugging only)
      const result = await response.json();
      if (result.errors) {
        console.log('Label creation error:', result.errors);
      }
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

  test('labels are displayed in create card modal', async ({ page }) => {
    await navigateToBoard(page);

    // Open create card modal
    await clickAddCardInColumn(page, 'Todo');
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });

    // Labels section should be visible with search input
    await expect(page.getByText('Labels', { exact: true })).toBeVisible();
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await expect(labelInput).toBeVisible();

    // Focus the input to show dropdown with existing labels
    await labelInput.focus();

    // Labels should appear in dropdown
    await expect(page.locator('.absolute.z-10').getByText('Bug')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.absolute.z-10').getByText('Feature')).toBeVisible();
    await expect(page.locator('.absolute.z-10').getByText('Enhancement')).toBeVisible();
    await expect(page.locator('.absolute.z-10').getByText('Documentation')).toBeVisible();

    // Press Escape to close dropdown and modal
    await page.keyboard.press('Escape');
    // Wait for dropdown to close then press again to close modal
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');
  });

  test('can create card with single label', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Bug Card ${randomId}`);

    // Select Bug label via dropdown
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    // Card should be created with label
    await expect(page.getByText(`Bug Card ${randomId}`)).toBeVisible({ timeout: 5000 });
  });

  test('can create card with multiple labels', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Multi Label Card ${randomId}`);

    // Select multiple labels via dropdown
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`Multi Label Card ${randomId}`)).toBeVisible({ timeout: 5000 });
  });

  test('selected labels have visual indicator', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    // Select Bug label via dropdown
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();

    // Bug should appear in dropdown
    const bugOption = page.locator('.absolute.z-10').getByText('Bug');
    await expect(bugOption).toBeVisible();

    // Select Bug
    await bugOption.click();

    // Selected label should appear as a badge above the input (span with text and X button)
    const selectedBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    await expect(selectedBadge).toBeVisible();

    // Close dropdown first, then click Cancel
    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');
  });

  test('can toggle labels on and off', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const labelInput = page.locator('input[placeholder*="search or create labels"]');

    // Select Bug label
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();

    // Should see Bug badge
    const bugBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    await expect(bugBadge).toBeVisible();

    // Click the X button inside the badge to remove
    await bugBadge.locator('button').click();

    // Bug badge should be gone
    await expect(bugBadge).not.toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('can add labels to existing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card without labels
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `Add Labels Later ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Add Labels Later ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and add labels
    await page.getByText(`Add Labels Later ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Add Enhancement label via dropdown
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Enhancement').click();

    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('can remove labels from existing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card with label
    await clickAddCardInColumn(page, 'Done');
    await page.fill('#title', `Remove Labels ${randomId}`);

    // Add Documentation label via dropdown
    const createLabelInput = page.locator('input[placeholder*="search or create labels"]');
    await createLabelInput.focus();
    await page.locator('.absolute.z-10').getByText('Documentation').click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Remove Labels ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and remove label
    await page.getByText(`Remove Labels ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Documentation should be visible as a badge, click X to remove
    const docBadge = page.locator('span.inline-flex').filter({ hasText: 'Documentation' }).filter({ has: page.locator('button') });
    await expect(docBadge).toBeVisible();
    await docBadge.locator('button').click();

    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('labels persist after editing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card with labels
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Persist Labels ${randomId}`);

    // Add labels via dropdown
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Persist Labels ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Edit the card (change title only)
    await page.getByText(`Persist Labels ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Verify labels are still visible as badges
    const bugBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    const featureBadge = page.locator('span.inline-flex').filter({ hasText: 'Feature' }).filter({ has: page.locator('button') });
    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();

    // Change title
    await page.fill('#title', `Persist Labels Updated ${randomId}`);
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Re-open and verify labels are still there
    await page.getByText(`Persist Labels Updated ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('labels display with correct colors', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    // Open dropdown to see labels with colors
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();

    // Labels in dropdown should have colored dots/indicators
    const dropdown = page.locator('.absolute.z-10');
    await expect(dropdown.getByText('Bug')).toBeVisible();
    await expect(dropdown.getByText('Feature')).toBeVisible();
    await expect(dropdown.getByText('Enhancement')).toBeVisible();

    // Select Bug to verify it appears with color
    await dropdown.getByText('Bug').click();

    // Bug badge should be visible with its color styling
    const bugBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    await expect(bugBadge).toBeVisible();
    const bugStyle = await bugBadge.getAttribute('style');
    expect(bugStyle).toContain('239, 68, 68');

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('card with all labels can be created', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `All Labels ${randomId}`);

    // Select all labels via dropdown
    const labelInput = page.locator('input[placeholder*="search or create labels"]');
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Enhancement').click();
    await labelInput.focus();
    await page.locator('.absolute.z-10').getByText('Documentation').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`All Labels ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Verify all labels are saved
    await page.getByText(`All Labels ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Verify all labels are visible as badges
    const bugBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    const featureBadge = page.locator('span.inline-flex').filter({ hasText: 'Feature' }).filter({ has: page.locator('button') });
    const enhancementBadge = page.locator('span.inline-flex').filter({ hasText: 'Enhancement' }).filter({ has: page.locator('button') });
    const docBadge = page.locator('span.inline-flex').filter({ hasText: 'Documentation' }).filter({ has: page.locator('button') });

    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();
    await expect(enhancementBadge).toBeVisible();
    await expect(docBadge).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });
});
