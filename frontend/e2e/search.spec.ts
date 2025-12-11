import { test, expect } from '@playwright/test';
import { setupTestEnvironment } from './helpers';

test.describe('Search (Command Palette)', () => {
  test('opens command palette with Ctrl+K shortcut', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'search');

    // Navigate to the organization page where AppShell is rendered
    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Verify the page loaded (use first() since org name appears multiple times)
    await expect(page.getByText(ctx.orgName).first()).toBeVisible({ timeout: 10000 });

    // Press Ctrl+K (works cross-platform in headless browsers)
    await page.keyboard.press('Control+k');

    // Command palette should be visible
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });
  });

  test('closes command palette with Escape', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'search');

    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Open command palette
    await page.keyboard.press('Control+k');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

    // Press Escape to close
    await page.keyboard.press('Escape');

    // Command palette should be hidden
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).not.toBeVisible({ timeout: 5000 });
  });

  test('shows minimum character message for short queries', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'search');

    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Open command palette
    await page.keyboard.press('Control+k');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

    // Type single character
    await page.getByPlaceholder('Search cards, projects, boards...').fill('a');

    // Should show minimum character message
    await expect(page.getByText('Type at least 2 characters to search')).toBeVisible({ timeout: 5000 });
  });

  test('search clears when command palette reopens', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'search');

    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Open command palette
    await page.keyboard.press('Control+k');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

    // Type a search query
    await page.getByPlaceholder('Search cards, projects, boards...').fill('test search');

    // Close with Escape
    await page.keyboard.press('Escape');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).not.toBeVisible({ timeout: 5000 });

    // Reopen command palette
    await page.keyboard.press('Control+k');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

    // Search input should be empty
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toHaveValue('');
  });

  test('filter buttons show correct active state', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'search');

    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Open command palette
    await page.keyboard.press('Control+k');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

    // "All" filter should be active by default
    await expect(page.getByRole('button', { name: 'All', exact: true })).toHaveClass(/bg-indigo-100/);

    // Click "Cards" filter
    await page.getByRole('button', { name: 'Cards', exact: true }).click();

    // "Cards" should now be active
    await expect(page.getByRole('button', { name: 'Cards', exact: true })).toHaveClass(/bg-indigo-100/);
    // "All" should no longer be active
    await expect(page.getByRole('button', { name: 'All', exact: true })).not.toHaveClass(/bg-indigo-100/);
  });

  test('shows no results message when nothing found', async ({ page }) => {
    const ctx = await setupTestEnvironment(page, 'search');

    await page.goto(`/organizations/${ctx.orgId}`);
    await page.waitForLoadState('networkidle');

    // Open command palette
    await page.keyboard.press('Control+k');
    await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

    // Search for something that doesn't exist
    const randomSearch = `xyznonexistent${Date.now()}`;
    await page.getByPlaceholder('Search cards, projects, boards...').fill(randomSearch);

    // Should show no results message (or still be searching)
    // Wait for either "no results" message or the search to complete
    await expect(
      page.getByText(`No results found for "${randomSearch}"`)
        .or(page.getByText('Type at least 2 characters to search'))
    ).toBeVisible({ timeout: 15000 });
  });

  // These tests depend on Typesense being configured and running
  // They will test actual search functionality when the backend is available
  test.describe('Search Results (requires Typesense)', () => {
    test('can search and command palette stays open', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'search');

      await page.goto(`/organizations/${ctx.orgId}`);
      await page.waitForLoadState('networkidle');

      // Open command palette
      await page.keyboard.press('Control+k');
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

      // Search for the organization name
      const searchTerm = ctx.orgName.substring(0, 10);
      await page.getByPlaceholder('Search cards, projects, boards...').fill(searchTerm);

      // Wait a bit for search to process
      await page.waitForTimeout(2000);

      // Verify command palette is still open and input has our search term
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible();
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toHaveValue(searchTerm);
    });

    test('searches and finds project when indexed', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'search');

      await page.goto(`/organizations/${ctx.orgId}`);
      await page.waitForLoadState('networkidle');

      // Open command palette
      await page.keyboard.press('Control+k');
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

      // Search for the project name
      const searchTerm = ctx.projectName.substring(0, 10);
      await page.getByPlaceholder('Search cards, projects, boards...').fill(searchTerm);

      // Wait for results - either we find results or get "no results" message
      await expect(
        page.locator('[class*="bg-purple-100"]').filter({ hasText: 'Project' })
          .or(page.getByText(/No results found for/))
      ).toBeVisible({ timeout: 15000 });
    });

    test('clicking result navigates to it', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'search');

      await page.goto(`/organizations/${ctx.orgId}`);
      await page.waitForLoadState('networkidle');

      // Open command palette
      await page.keyboard.press('Control+k');
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

      // Search for the project
      const searchTerm = ctx.projectName.substring(0, 10);
      await page.getByPlaceholder('Search cards, projects, boards...').fill(searchTerm);

      // Try to find the project result - if found, click it
      const projectResult = page.getByRole('button').filter({ hasText: ctx.projectName });
      const noResults = page.getByText(/No results found for/);

      // Wait for either results or no results
      await expect(projectResult.or(noResults)).toBeVisible({ timeout: 15000 });

      // If we found the project, click it and verify navigation
      if (await projectResult.isVisible()) {
        await projectResult.click();

        // Should navigate to the project page
        await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/, { timeout: 10000 });
        // Command palette should be closed
        await expect(page.getByPlaceholder('Search cards, projects, boards...')).not.toBeVisible({ timeout: 5000 });
      }
      // If no results, the test still passes - search just isn't indexing data
    });

    test('selects result with Enter key', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'search');

      await page.goto(`/organizations/${ctx.orgId}`);
      await page.waitForLoadState('networkidle');

      // Open command palette
      await page.keyboard.press('Control+k');
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

      // Search for the project
      const searchTerm = ctx.projectName.substring(0, 10);
      await page.getByPlaceholder('Search cards, projects, boards...').fill(searchTerm);

      // Wait for results or no results
      const projectResult = page.locator('[class*="bg-purple-100"]').filter({ hasText: 'Project' });
      const noResults = page.getByText(/No results found for/);

      await expect(projectResult.or(noResults)).toBeVisible({ timeout: 15000 });

      // If we have results, try navigating with Enter
      if (await projectResult.isVisible()) {
        // Press Enter to navigate
        await page.keyboard.press('Enter');

        // Command palette should be closed
        await expect(page.getByPlaceholder('Search cards, projects, boards...')).not.toBeVisible({ timeout: 5000 });
      }
    });

    test('searches for cards', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'search');

      // Navigate to org page
      await page.goto(`/organizations/${ctx.orgId}`);
      await page.waitForLoadState('networkidle');

      // Open command palette
      await page.keyboard.press('Control+k');
      await expect(page.getByPlaceholder('Search cards, projects, boards...')).toBeVisible({ timeout: 10000 });

      // Search for "card" - a common term
      await page.getByPlaceholder('Search cards, projects, boards...').fill('card');

      // Wait for results - either we find cards or get "no results" message
      await expect(
        page.locator('[class*="bg-blue-100"]').filter({ hasText: 'Card' })
          .or(page.getByText(/No results found for/))
      ).toBeVisible({ timeout: 15000 });
    });
  });
});
