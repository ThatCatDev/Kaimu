import { test, expect } from '@playwright/test';
import { setupTestEnvironment, navigateToBoard, getColumn, clickAddCardInColumn } from './helpers';

test.describe('Kanban Cards with Tags', () => {
  // Helper function to create tags via GraphQL
  async function createTagsForProject(page: any, projectId: string, tags: Array<{ name: string, color: string, description: string }>) {
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
      const result = await response.json();
      if (result.errors) {
        console.log('Tag creation error:', result.errors);
      }
    }
  }

  test('tags are displayed in create card modal', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    // Create tags for this test
    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
      { name: 'Documentation', color: '#8B5CF6', description: 'Documentation needs update' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    // Open create card modal
    await clickAddCardInColumn(page, 'Todo');

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Bug Card ${ctx.testId}`);

    // Select Bug tag via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    // Card should be created with tag
    await expect(page.getByText(`Bug Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
  });

  test('can create card with multiple tags', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Multi Tag Card ${ctx.testId}`);

    // Select multiple tags via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();

    await page.getByRole('button', { name: 'Create Card' }).click();

    await expect(page.getByText(`Multi Tag Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });
  });

  test('selected tags have visual indicator', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
    ]);

    await navigateToBoard(page, ctx.projectId);

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
    ]);

    await navigateToBoard(page, ctx.projectId);

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    // Create card without tags
    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `Add Tags Later ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Add Tags Later ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and add tags
    await page.getByText(`Add Tags Later ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Add Enhancement tag via dropdown - use the detail panel dialog
    const detailDialog = page.getByRole('dialog', { name: 'Card Details' });
    const tagInput = detailDialog.getByPlaceholder('Type to search or create tags');
    await tagInput.click();
    // Wait for dropdown to appear and stabilize
    await page.waitForTimeout(300);
    // Use more specific selector for the dropdown that's visible
    await detailDialog.locator('.absolute.z-10').getByText('Enhancement').click();

    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('can remove tags from existing card', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Documentation', color: '#8B5CF6', description: 'Documentation needs update' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    // Create card with tag
    await clickAddCardInColumn(page, 'Done');
    await page.fill('#title', `Remove Tags ${ctx.testId}`);

    // Add Documentation tag via dropdown
    const createTagInput = page.locator('input[placeholder*="search or create tags"]');
    await createTagInput.focus();
    await page.locator('.absolute.z-10').getByText('Documentation').click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Remove Tags ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail and remove tag
    await page.getByText(`Remove Tags ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Documentation should be visible as a badge in Card Details dialog, click X to remove
    const dialog = page.getByRole('dialog', { name: 'Card Details' });
    const docBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Documentation' }).filter({ has: page.locator('button') });
    await expect(docBadge).toBeVisible();
    await docBadge.locator('button').click();

    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });
  });

  test('tags persist after editing card', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    // Create card with tags
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Persist Tags ${ctx.testId}`);

    // Add tags via dropdown
    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Bug').click();
    await tagInput.focus();
    await page.locator('.absolute.z-10').getByText('Feature').click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Persist Tags ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Edit the card (change title only)
    await page.getByText(`Persist Tags ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Verify tags are still visible as badges (scoped to Card Details dialog)
    const dialog = page.getByRole('dialog', { name: 'Card Details' });
    const bugBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    const featureBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Feature' }).filter({ has: page.locator('button') });
    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();

    // Change title (scoped to dialog) - CardDetailModal uses "detail-" prefix
    await dialog.locator('#detail-title').fill(`Persist Tags Updated ${ctx.testId}`);
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Close' }).click();

    // Re-open and verify tags are still there
    await page.getByText(`Persist Tags Updated ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });
    // Re-scope to the new dialog
    const dialog2 = page.getByRole('dialog', { name: 'Card Details' });
    await expect(dialog2.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') })).toBeVisible();
    await expect(dialog2.locator('span.inline-flex').filter({ hasText: 'Feature' }).filter({ has: page.locator('button') })).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('tags display with correct colors', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
    ]);

    await navigateToBoard(page, ctx.projectId);

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
      { name: 'Documentation', color: '#8B5CF6', description: 'Documentation needs update' },
    ]);

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'In Progress');
    await page.fill('#title', `All Tags ${ctx.testId}`);

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

    await expect(page.getByText(`All Tags ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Verify all tags are saved
    await page.getByText(`All Tags ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Verify all tags are visible as badges (scoped to Card Details dialog)
    const dialog = page.getByRole('dialog', { name: 'Card Details' });
    const bugBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Bug' }).filter({ has: page.locator('button') });
    const featureBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Feature' }).filter({ has: page.locator('button') });
    const enhancementBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Enhancement' }).filter({ has: page.locator('button') });
    const docBadge = dialog.locator('span.inline-flex').filter({ hasText: 'Documentation' }).filter({ has: page.locator('button') });

    await expect(bugBadge).toBeVisible();
    await expect(featureBadge).toBeVisible();
    await expect(enhancementBadge).toBeVisible();
    await expect(docBadge).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('can filter tags by typing', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
      { name: 'Feature', color: '#10B981', description: 'New functionality' },
      { name: 'Enhancement', color: '#3B82F6', description: 'Improvement to existing feature' },
    ]);

    await navigateToBoard(page, ctx.projectId);

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await expect(tagInput).toBeVisible({ timeout: 5000 });
    await tagInput.focus();

    // Type a tag name that doesn't exist
    const newTagName = `NewTag${ctx.testId}`;
    await tagInput.fill(newTagName);
    await page.waitForTimeout(300); // Wait for dropdown animation

    // Should show "Create" option in dropdown
    const createOption = page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`);
    await expect(createOption).toBeVisible({ timeout: 5000 });

    await page.keyboard.press('Escape');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');
  });

  test('can create new tag with color picker from create card modal', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Card With New Tag ${ctx.testId}`);

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a new tag name
    const newTagName = `CustomTag${ctx.testId}`;
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

    // New tag should appear as a badge in Create Card modal
    const createCardModal = page.getByLabel('Create Card');
    const newTagBadge = createCardModal.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible();

    // Create the card
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Card With New Tag ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Verify the new tag persists
    await page.getByText(`Card With New Tag ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Check badge in Card Details modal
    const cardDetailsModal = page.getByLabel('Card Details');
    const newTagBadgeInDetails = cardDetailsModal.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadgeInDetails).toBeVisible();

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('can create new tag using Enter key', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a new tag name and press Enter
    const newTagName = `EnterTag${ctx.testId}`;
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
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    // Type a new tag name
    const newTagName = `CancelTag${ctx.testId}`;
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
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `PreviewTag${ctx.testId}`;
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
    const ctx = await setupTestEnvironment(page, 'tags');

    await createTagsForProject(page, ctx.projectId, [
      { name: 'Bug', color: '#EF4444', description: 'Something is broken' },
    ]);

    await navigateToBoard(page, ctx.projectId);

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `EscTag${ctx.testId}`;
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
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    // Create a new tag
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `First Card ${ctx.testId}`);

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `SharedTag${ctx.testId}`;
    await tagInput.fill(newTagName);
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    const colorPicker = page.locator('.absolute.z-20');
    await colorPicker.getByRole('button', { name: 'Create Tag' }).click();

    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`First Card ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

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
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    // Create a card without tags
    await clickAddCardInColumn(page, 'Todo');
    await page.fill('#title', `Edit Mode Tag ${ctx.testId}`);
    await page.getByRole('button', { name: 'Create Card' }).click();
    await expect(page.getByText(`Edit Mode Tag ${ctx.testId}`)).toBeVisible({ timeout: 5000 });

    // Open card detail
    await page.getByText(`Edit Mode Tag ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Create a new tag from edit mode - scope to Card Details dialog to avoid Create Card dialog
    const cardDetailsDialog = page.getByRole('dialog', { name: 'Card Details' });
    const tagInput = cardDetailsDialog.getByPlaceholder('Type to search or create tags');
    await tagInput.focus();

    const newTagName = `EditModeTag${ctx.testId}`;
    await tagInput.fill(newTagName);
    await page.locator('.absolute.z-10').getByText(`Create "${newTagName}"`).click();

    // Color picker should appear
    const colorPicker = page.locator('.absolute.z-20');
    await expect(colorPicker).toBeVisible({ timeout: 5000 });

    // Select a color and create
    await colorPicker.locator('button[style*="background-color: rgb(20, 184, 166)"]').click(); // teal
    await colorPicker.getByRole('button', { name: 'Create Tag' }).click();

    // Tag should be added (scope to Card Details modal to avoid conflicts)
    const cardDetailsModal = page.getByLabel('Card Details');
    let newTagBadge = cardDetailsModal.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible({ timeout: 5000 });

    // Wait for auto-save
    await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).not.toBeVisible({ timeout: 5000 });

    // Reopen and verify tag persisted
    await page.getByText(`Edit Mode Tag ${ctx.testId}`).click();
    await expect(page.getByRole('heading', { name: 'Card Details' })).toBeVisible({ timeout: 5000 });

    // Re-locate the badge after reopening (scope to Card Details modal)
    const cardDetailsModal2 = page.getByLabel('Card Details');
    newTagBadge = cardDetailsModal2.locator('span.inline-flex').filter({ hasText: newTagName }).filter({ has: page.locator('button') });
    await expect(newTagBadge).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Close' }).click();
  });

  test('selected color in picker has visual indicator', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'tags');

    await navigateToBoard(page, ctx.projectId);

    await clickAddCardInColumn(page, 'Todo');

    const tagInput = page.locator('input[placeholder*="search or create tags"]');
    await tagInput.focus();

    const newTagName = `IndicatorTest${ctx.testId}`;
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
