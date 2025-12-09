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

test.describe('Kanban Cards with Tags', () => {
  // Generate unique identifiers for this test run
  const randomId = Math.random().toString(36).substring(2, 10);
  const testUser = `tags_e2e_${randomId}`;
  const password = 'testpassword123';
  let organizationId: string;
  let projectId: string;
  const orgName = `Tags Test Org ${randomId}`;
  const projectName = `Tags Test Project ${randomId}`;
  const projectKey = `TG${randomLetters(4)}`;

  test.beforeAll(async ({ browser }) => {
    // Register a user, create an organization, project, and tags
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

    // Create tags via GraphQL API using page's request context (has auth cookies)
    const tags = [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
      { name: 'Documentation', color: '#8B5CF6', description: 'Documentation needs update' },
    ];

    for (const tag of tags) {
      const response = await page.request.post('http://localhost:3000/graphql', {
        data: {
          query: `
            mutation CreateTag($input: CreateTagInput!) {
              createTag(input: $input) {
                id
                name
                color
              }
            }
          `,
          variables: {
            input: {
              projectId,
              name: tag.name,
              color: tag.color,
              description: tag.description,
            },
          },
        },
      });
      // Handle response (log errors for debugging only)
      const result = await response.json();
      if (result.errors) {
        console.log('Tag creation error:', result.errors);
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

  test('tags are displayed in create card modal', async ({ page }) => {
    await navigateToBoard(page);

    // Open create card modal
    await clickAddCardInColumn(page, 'Todo');
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });

    // Tags section should be visible with search input
    await expect(page.getByText('Tags', { exact: true })).toBeVisible();
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await expect(tagInput).toBeVisible();

    // Focus the input to show dropdown with existing tags
    await tagInput.focus();

    // Tags should appear in dropdown
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

  test('can create card with single tag', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Bug Card ${randomId}`);

    // Select Bug tag via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    // Card should be created with tag
    await expect(page.getByText(`Bug Card ${randomId}`)).toBeVisible({ timeout: 5000 });
  });

  test('can create card with multiple tags', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Multi Tag Card ${randomId}`);

    // Select multiple tags via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`Multi Tag Card ${randomId}`)).toBeVisible({ timeout: 5000 });
  });

  test('selected tags have visual indicator', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    // Select Bug tag via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Bug should appear in dropdown
    const bugOption = page.locator('.absolute.z-10').getByText('Bug');
    await expect(bugOption).toBeVisible();

    // Select Bug
    await bugOption.click();

    // Selected tag should appear as a badge above the input (span with text and X button)
    const selectedBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    await expect(selectedBadge).toBeVisible();

    // Close dropdown first, then click Cancel
    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');
  });

  test('can toggle tags on and off', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');

    // Select Bug tag
    await tagInput.focus();
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

  test('can add tags to existing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card without tags
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `Add Tags Later ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Add Tags Later ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and add tags
    await page.getByText(`Add Tags Later ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Add Enhancement tag via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Enhancement').click();

    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('can remove tags from existing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card with tag
    await clickAddCardInColumn(page, 'Done');
    await page.fill('#title', `Remove Tags ${randomId}`);

    // Add Documentation tag via dropdown
    const createTagInput = page.locator('input[placeholder*="search or create tags"]');
    await createTagInput.focus();
    await page.locator('.absolute.z-10').getByText('Documentation').click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Remove Tags ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and remove tag
    await page.getByText(`Remove Tags ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Documentation should be visible as a badge, click X to remove
    const docBadge = page.locator('span.inline-flex').filter({ hasText: 'Documentation' }).filter({ has: page.locator('button') });
    await expect(docBadge).toBeVisible();
    await docBadge.locator('button').click();

    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('tags persist after editing card', async ({ page }) => {
    await navigateToBoard(page);

    // Create card with tags
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Persist Tags ${randomId}`);

    // Add tags via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Persist Tags ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Edit the card (change title only)
    await page.getByText(`Persist Tags ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Verify tags are still visible as badges
    const bugBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    const featureBadge = page.locator('span.inline-flex').filter({ hasText: 'Feature' }).filter({ has: page.locator('button') });
    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();

    // Change title
    await page.fill('#title', `Persist Tags Updated ${randomId}`);
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Re-open and verify tags are still there
    await page.getByText(`Persist Tags Updated ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('tags display with correct colors', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    // Open dropdown to see tags with colors
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Tags in dropdown should have colored dots/indicators
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

  test('card with all tags can be created', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `All Tags ${randomId}`);

    // Select all tags via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Enhancement').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Documentation').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`All Tags ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Verify all tags are saved
    await page.getByText(`All Tags ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Verify all tags are visible as badges
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

  test('can filter tags by typing', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // All tags should be visible initially
    await expect(page.locator('.absolute.z-10').getByText('Bug')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.absolute.z-10').getByText('Feature')).toBeVisible();

    // Type to filter
    await tagInput.fill('Bug');

    // Only Bug should be visible now
    await expect(page.locator('.absolute.z-10').getByText('Bug')).toBeVisible();
    await expect(page.locator('.absolute.z-10').getByText('Feature')).not.toBeVisible();
    await expect(page.locator('.absolute.z-10').getByText('Enhancement')).not.toBeVisible();

    // Clear and type different filter
    await tagInput.fill('Feat');
    await expect(page.locator('.absolute.z-10').getByText('Feature')).toBeVisible();
    await expect(page.locator('.absolute.z-10').getByText('Bug')).not.toBeVisible();

    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');
  });

  test('shows create tag option when typing non-existing tag', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a tag name that doesn't exist
    const newTagName = `NewTag${randomId}`;
    await tagInput.fill(newTagName);

    // Should show "Create" option in dropdown
    const createOption = page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`);
    await expect(createOption).toBeVisible({ timeout: 5000 });

    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');
  });

  test('can create new tag with color picker from create card modal', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Card With New Tag ${randomId}`);

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a new tag name
    const newTagName = `CustomTag${randomId}`;
    await tagInput.fill(newTagName);

    // Click "Create" option to open color picker
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    // Color picker should appear
    await expect(page.getByText(`Choose color for "${newTagName}"`)).toBeVisible({ timeout: 5000 });

    // Should see color grid with preset colors
    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible();

    // Should see preview of the tag (use exact match to avoid matching the header text)
    await expect(colorPicker.getByText(newTagName, { exact: true })).toBeVisible();

    // Click a specific color (green - #22c55e)
    await colorPicker.locator('button[style*="background-color: rgb(34, 197, 94)"]').click();

    // Click "Create Tag" button
    await colorPicker.getByRole('button', { name: 'Create Tag' }).click();

    // Color picker should close and tag should be selected
    await expect(colorPicker).not.toBeVisible({ timeout: 5000 });

    // New tag should appear as a badge
    const newTagBadge = page.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible();

    // Create the card
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Card With New Tag ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Verify the new tag persists
    await page.getByText(`Card With New Tag ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    await expect(newTagBadge).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('can create new tag using Enter key', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a new tag name and press Enter
    const newTagName = `EnterTag${randomId}`;
    await tagInput.fill(newTagName);
    await page.keyboard.press('Enter');

    // Color picker should appear
    await expect(page.getByText(`Choose color for "${newTagName}"`)).toBeVisible({ timeout: 5000 });

    // Create with default color
    await page.locator('.absolute.z-20').getByRole('button', { name: 'Create Tag' }).click();

    // New tag should appear as a badge
    const newTagBadge = page.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('can cancel color picker without creating tag', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a new tag name
    const newTagName = `CancelTag${randomId}`;
    await tagInput.fill(newTagName);

    // Click create to open color picker
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    // Color picker should appear
    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible({ timeout: 5000 });

    // Click Cancel
    await colorPicker.getByRole('button', { name: 'Cancel' }).click();

    // Color picker should close
    await expect(colorPicker).not.toBeVisible({ timeout: 5000 });

    // Tag should NOT be created (no badge)
    const tagBadge = page.locator('span.inline-flex').filter({ hasText: newTagName });
    await expect(tagBadge).not.toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('color picker shows preview that updates when selecting colors', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `PreviewTag${randomId}`;
    await tagInput.fill(newTagName);
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible({ timeout: 5000 });

    // Get the preview element
    const preview = colorPicker.locator('span.inline-flex').filter({ hasText: newTagName });
    await expect(preview).toBeVisible();

    // Click red color (#ef4444)
    await colorPicker.locator('button[style*="background-color: rgb(239, 68, 68)"]').click();

    // Preview should have red styling
    const previewStyle = await preview.getAttribute('style');
    expect(previewStyle).toContain('239, 68, 68');

    // Click blue color (#3b82f6)
    await colorPicker.locator('button[style*="background-color: rgb(59, 130, 246)"]').click();

    // Preview should now have blue styling
    const newPreviewStyle = await preview.getAttribute('style');
    expect(newPreviewStyle).toContain('59, 130, 246');

    await colorPicker.getByRole('button', { name: 'Cancel' }).click();
    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('can edit existing tag color by clicking on badge', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Select Bug tag
    await page.locator('.absolute.z-10').getByText('Bug').click();

    // Bug badge should be visible
    const bugBadge = page.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    await expect(bugBadge).toBeVisible();

    // Click on the badge text area (not the X button) to edit color
    await bugBadge.click();

    // Color picker should appear with "Edit color for" text
    await expect(page.getByText('Edit color for "Bug"')).toBeVisible({ timeout: 5000 });

    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible();

    // Select a different color (purple - #8b5cf6)
    await colorPicker.locator('button[style*="background-color: rgb(139, 92, 246)"]').click();

    // Click "Save Color" button
    await colorPicker.getByRole('button', { name: 'Save Color' }).click();

    // Color picker should close
    await expect(colorPicker).not.toBeVisible({ timeout: 5000 });

    // Bug badge should still be visible (tag not removed)
    await expect(bugBadge).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('color picker closes with Escape key from tag input', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `EscTag${randomId}`;
    await tagInput.fill(newTagName);

    // Press Enter to open color picker (keeps focus on input where Escape handler works)
    await page.keyboard.press('Enter');

    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible({ timeout: 5000 });

    // Focus the tag input again so the TagPicker's keydown handler catches Escape
    await tagInput.focus();

    // Press Escape to close color picker
    await page.keyboard.press('Escape');

    // Color picker should close but modal should stay open
    await expect(colorPicker).not.toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('newly created tag appears in dropdown for future cards', async ({ page }) => {
    await navigateToBoard(page);

    // Create a new tag
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `First Card ${randomId}`);

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `SharedTag${randomId}`;
    await tagInput.fill(newTagName);
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    const colorPicker = page.locator('.absolute.z-20');
    await colorPicker.getByRole('button', { name: 'Create Tag' }).click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`First Card ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Create another card and verify the tag is available
    await clickAddCardInColumn(page, 'Todo');

    const tagInput2 = page.locator('input[placeholder*="search or create tags"]');
    await tagInput2.focus();

    // The new tag should appear in the dropdown
    await expect(page.locator('.absolute.z-10').getByText(newTagName)).toBeVisible({ timeout: 5000 });

    // Can select it
    await page.locator('.absolute.z-10').getByText(newTagName).click();

    const tagBadge = page.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(tagBadge).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('can create tag from card detail modal (edit mode)', async ({ page }) => {
    await navigateToBoard(page);

    // Create a card without tags
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Edit Mode Tag ${randomId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Edit Mode Tag ${randomId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail
    await page.getByText(`Edit Mode Tag ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Create a new tag from edit mode
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `EditModeTag${randomId}`;
    await tagInput.fill(newTagName);
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    // Color picker should appear
    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible({ timeout: 5000 });

    // Select a color and create
    await colorPicker.locator('button[style*="background-color: rgb(20, 184, 166)"]').click(); // teal
    await colorPicker.getByRole('button', { name: 'Create Tag' }).click();

    // Tag should be added
    let newTagBadge = page.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible({ timeout: 5000 });

    // Wait for auto-save
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Reopen and verify tag persisted
    await page.getByText(`Edit Mode Tag ${randomId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Re-locate the badge after reopening
    newTagBadge = page.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('selected color in picker has visual indicator', async ({ page }) => {
    await navigateToBoard(page);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `IndicatorTest${randomId}`;
    await tagInput.fill(newTagName);
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible({ timeout: 5000 });

    // Click on a color
    const greenButton = colorPicker.locator('button[style*="background-color: rgb(34, 197, 94)"]');
    await greenButton.click();

    // The selected color button should have a visual indicator (border or ring)
    const greenButtonClass = await greenButton.getAttribute('class');
    expect(greenButtonClass).toContain('border-gray-800');

    // Click another color
    const blueButton = colorPicker.locator('button[style*="background-color: rgb(59, 130, 246)"]');
    await blueButton.click();

    // Blue should now have the indicator
    const blueButtonClass = await blueButton.getAttribute('class');
    expect(blueButtonClass).toContain('border-gray-800');

    // Green should no longer have the indicator
    const greenButtonClassAfter = await greenButton.getAttribute('class');
    expect(greenButtonClassAfter).not.toContain('border-gray-800');

    await colorPicker.getByRole('button', { name: 'Cancel' }).click();
    await page.getByRole('button', { name: 'Cancel' }).click();
  });
});
