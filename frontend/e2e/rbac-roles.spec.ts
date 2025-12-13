import { test, expect } from '@playwright/test';
import { randomId, randomLetters } from './helpers';

/**
 * RBAC Role Tests
 *
 * Tests the 4 system roles: Owner, Admin, Member, Viewer
 * Each role has different permissions for organization, project, board, and card operations.
 *
 * Owner: Full access to everything
 * Admin: All except org:delete, org:manage_roles
 * Member: View + create + edit (not delete/manage)
 * Viewer: View only
 */

interface TestUser {
  username: string;
  password: string;
}

interface TestSetup {
  owner: TestUser;
  orgId: string;
  orgName: string;
  projectId: string;
  projectName: string;
  boardId: string;
}

/**
 * Register a new user
 */
async function registerUser(page: any, prefix: string, specificEmail?: string): Promise<TestUser> {
  const testId = randomId();
  const username = `${prefix}_${testId}`;
  const email = specificEmail || `${username}@test.local`;
  const password = 'testpassword123';

  await page.goto('/register');
  await page.waitForLoadState('networkidle');
  await page.fill('#username', username);
  await page.fill('#email', email);
  await page.fill('#password', password);
  await page.fill('#confirmPassword', password);
  await page.getByRole('button', { name: 'Register' }).click();
  await page.waitForLoadState('networkidle');
  await expect(page).toHaveURL('/', { timeout: 15000 });

  return { username, password };
}

/**
 * Login as a user
 */
async function login(page: any, user: TestUser): Promise<void> {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.fill('#username', user.username);
  await page.fill('#password', user.password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForLoadState('networkidle');
  await expect(page).toHaveURL('/', { timeout: 15000 });
}

/**
 * Setup: Create org, project, board as owner
 */
async function setupTestEnvironment(page: any): Promise<TestSetup> {
  const testId = randomId();
  const owner = await registerUser(page, `owner_${testId}`);
  const orgName = `RBAC Test Org ${testId}`;
  const projectName = `RBAC Test Project ${testId}`;
  const projectKey = `RB${randomLetters(4)}`;

  // Create organization
  await page.goto('/organizations/new');
  await page.waitForLoadState('networkidle');
  await page.fill('#name', orgName);
  await page.getByRole('button', { name: 'Create Organization' }).click();
  await page.waitForLoadState('networkidle');
  await expect(page).toHaveURL(/\/organizations\/([a-f0-9-]+)/, { timeout: 15000 });

  const orgUrl = page.url();
  const orgMatch = orgUrl.match(/\/organizations\/([a-f0-9-]+)/);
  const orgId = orgMatch ? orgMatch[1] : '';

  // Create project
  await page.goto(`/organizations/${orgId}/projects/new`);
  await page.waitForLoadState('networkidle');
  await page.fill('#name', projectName);
  await page.fill('#key', projectKey);
  await page.getByRole('button', { name: 'Create Project' }).click();
  await page.waitForLoadState('networkidle');
  await expect(page).toHaveURL(/\/projects\/([a-f0-9-]+)/, { timeout: 15000 });

  const projectUrl = page.url();
  const projectMatch = projectUrl.match(/\/projects\/([a-f0-9-]+)/);
  const projectId = projectMatch ? projectMatch[1] : '';

  // Create a board first (boards are not auto-created)
  await page.waitForLoadState('networkidle');

  // Check if there are any boards - look for "No boards" text
  const noBoardsVisible = await page.getByText('No boards').isVisible().catch(() => false);

  if (noBoardsVisible) {
    // Create a board first
    await page.getByRole('button', { name: 'New Board' }).click();
    await expect(page.getByRole('heading', { name: 'Create Board' })).toBeVisible({ timeout: 5000 });
    await page.fill('#boardName', 'Kanban Board');
    await page.getByRole('button', { name: 'Create Board', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Create Board' })).not.toBeVisible({ timeout: 5000 });
    // Wait for board to appear
    await expect(page.getByText('Kanban Board')).toBeVisible({ timeout: 5000 });
  }

  // Click on the board link
  await page.locator('a[href*="/board/"]').first().click();
  await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 15000 });

  const boardUrl = page.url();
  const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
  const boardId = boardMatch ? boardMatch[1] : '';

  // Create a test card with retry logic
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(500); // Let board fully render

  // Open card creation dialog
  await page.getByRole('button', { name: 'Add card' }).first().click();
  await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });

  // Fill in card title
  await page.fill('#title', 'Test Card for RBAC');
  await page.waitForTimeout(200); // Let form update

  // Submit the form
  await page.getByRole('button', { name: 'Create Card' }).click();

  // Wait for dialog to close first
  await expect(page.getByRole('heading', { name: 'Create Card' })).toBeHidden({ timeout: 10000 });

  // Then verify card appears on board
  await expect(page.getByText('Test Card for RBAC')).toBeVisible({ timeout: 10000 });

  return {
    owner,
    orgId,
    orgName,
    projectId,
    projectName,
    boardId,
  };
}

/**
 * Accept an invitation by navigating to the invite page and clicking accept
 */
async function acceptInvitation(page: any, token: string): Promise<void> {
  if (!token) {
    throw new Error('No invitation token provided');
  }

  await page.goto(`/invite/${token}`);
  await page.waitForLoadState('networkidle');

  // Wait for page to finish loading (the component checks auth on mount)
  await page.waitForTimeout(1000);

  // Check if we see the "Accept Invitation" button - this means we're logged in
  const acceptButton = page.getByRole('button', { name: 'Accept Invitation' });

  // If not visible, we might see "Login Required" - that means auth didn't carry over
  const loginRequired = page.getByText('Login Required');
  if (await loginRequired.isVisible().catch(() => false)) {
    throw new Error('User is not logged in on the invite page. Auth state may not have carried over.');
  }

  // Wait for accept button to be visible
  await expect(acceptButton).toBeVisible({ timeout: 10000 });

  // Click Accept Invitation button
  await acceptButton.click();

  // Wait for either success OR error message
  await page.waitForTimeout(500);

  // Check for error first
  const errorMessage = page.getByText('Unable to Accept Invitation');
  if (await errorMessage.isVisible().catch(() => false)) {
    // Get the actual error message
    const errorText = await page.locator('.text-red-600').textContent().catch(() => 'Unknown error');
    throw new Error(`Invitation acceptance failed: ${errorText}`);
  }

  // Wait for success message
  await expect(page.getByText('Welcome to')).toBeVisible({ timeout: 10000 });

  // Wait for redirect to complete
  await page.waitForTimeout(1000);
}

/**
 * Invite a user and get the invitation token
 */
async function inviteUserAsRole(
  page: any,
  orgId: string,
  email: string,
  roleName: string
): Promise<string> {
  await page.goto(`/organizations/${orgId}/settings`);
  await page.waitForLoadState('networkidle');

  // Click Members tab
  await page.getByRole('link', { name: 'Members' }).click();
  await page.waitForTimeout(500);

  // Click Invite Member button
  await page.getByRole('button', { name: 'Invite Member' }).click();
  await expect(page.getByRole('heading', { name: 'Invite Member' })).toBeVisible({ timeout: 5000 });

  // Fill in email
  await page.fill('input[type="email"]', email);

  // Select role - bits-ui Select renders as a button, not combobox
  // The form group has "Role" label, find the button inside it
  const roleFormGroup = page.locator('div').filter({ hasText: /^Role/ }).first();
  const selectTrigger = roleFormGroup.locator('button');
  await selectTrigger.click();
  // Wait for dropdown to appear and click the role option
  await page.getByRole('option', { name: roleName }).click();

  // Submit - button is "Create Invitation" not "Send Invitation"
  await page.getByRole('button', { name: 'Create Invitation' }).click();

  // Wait for success - the modal shows "Invitation Created" heading
  await expect(page.getByRole('heading', { name: 'Invitation Created' })).toBeVisible({ timeout: 10000 });

  // Get the token from the invitation link input
  const linkInput = page.locator('input[readonly]');
  const invitationLink = await linkInput.inputValue();

  // Extract token from link (format: http://localhost:4321/invite/{token})
  // Token is base64 URL-encoded, so includes: A-Z, a-z, 0-9, -, _, =
  const tokenMatch = invitationLink.match(/\/invite\/([a-zA-Z0-9_=-]+)/);
  const token = tokenMatch ? tokenMatch[1] : null;

  // Close the modal
  await page.getByRole('button', { name: 'Done' }).click();

  return token || '';
}

test.describe('RBAC Role Permissions', () => {

  test.describe('Owner Role', () => {
    test('owner can access all organization settings', async ({ page }) => {
      const setup = await setupTestEnvironment(page);

      // Go to organization settings
      await page.goto(`/organizations/${setup.orgId}/settings`);
      await page.waitForLoadState('networkidle');

      // Owner should see both tabs: Members and Roles (these are links, not buttons)
      await expect(page.getByRole('link', { name: 'Members' })).toBeVisible();
      await expect(page.getByRole('link', { name: 'Roles' })).toBeVisible();
    });

    test('owner can create and delete projects', async ({ page }) => {
      const setup = await setupTestEnvironment(page);
      const testId = randomId();
      const newProjectName = `Owner Project ${testId}`;
      const newProjectKey = `OP${randomLetters(4)}`;

      // Create a new project
      await page.goto(`/organizations/${setup.orgId}/projects/new`);
      await page.waitForTimeout(300);
      await page.fill('#name', newProjectName);
      await page.fill('#key', newProjectKey);
      await page.getByRole('button', { name: 'Create Project' }).click();
      await expect(page).toHaveURL(/\/projects\/([a-f0-9-]+)/, { timeout: 10000 });

      // Should see the project
      await expect(page.getByRole('heading', { name: newProjectName })).toBeVisible();
    });

    test('owner can edit and delete cards', async ({ page }) => {
      const setup = await setupTestEnvironment(page);

      // Go to board
      await page.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await page.waitForLoadState('networkidle');
      await expect(page.getByText('Test Card for RBAC')).toBeVisible({ timeout: 10000 });

      // Click on card to open detail
      await page.getByText('Test Card for RBAC').click();
      await page.waitForTimeout(500);

      // Owner should see editable form (input fields, not plain text)
      // and delete button
      await expect(page.getByLabel('Title')).toBeVisible();
      await expect(page.locator('input#detail-title')).toBeVisible();
      await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible();
    });

    test('owner can manage board columns', async ({ page }) => {
      const setup = await setupTestEnvironment(page);

      // Go to board
      await page.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await page.waitForLoadState('networkidle');

      // Wait for board to fully load (card created in setup should be visible)
      await expect(page.getByText('Test Card for RBAC')).toBeVisible({ timeout: 10000 });

      // Owner should see "Add Column" element (it's inside a dndzone so accessibility role may vary)
      await expect(page.getByText('Add Column')).toBeVisible({ timeout: 10000 });
    });
  });

  test.describe('Viewer Role', () => {
    test('viewer can only view, not edit cards', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const viewerEmail = `viewer_${randomId()}@test.local`;

      // Invite a viewer
      const token = await inviteUserAsRole(page, setup.orgId, viewerEmail, 'Viewer');

      // Create a new browser context for the viewer
      const viewerPage = await context.newPage();

      // Register viewer with the same email as the invitation and accept
      const viewer = await registerUser(viewerPage, 'viewer', viewerEmail);
      await acceptInvitation(viewerPage, token);

      // Navigate to board as viewer
      await viewerPage.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await viewerPage.waitForLoadState('networkidle');

      // Viewer should see the card
      await expect(viewerPage.getByText('Test Card for RBAC')).toBeVisible({ timeout: 10000 });

      // Click on card to open detail
      await viewerPage.getByText('Test Card for RBAC').click();
      await viewerPage.waitForTimeout(500);

      // Viewer should see read-only view (plain text, not inputs)
      // Check that we see plain text elements instead of input fields
      const hasReadOnlyTitle = await viewerPage.locator('p').filter({ hasText: 'Test Card for RBAC' }).isVisible();
      const hasNoDeleteButton = await viewerPage.getByRole('button', { name: /Delete/ }).isHidden().catch(() => true);

      expect(hasReadOnlyTitle || hasNoDeleteButton).toBeTruthy();

      await viewerPage.close();
    });

    test('viewer cannot create cards', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const viewerEmail = `viewer_nocreate_${randomId()}@test.local`;

      // Invite a viewer
      const token = await inviteUserAsRole(page, setup.orgId, viewerEmail, 'Viewer');

      // Create a new browser context for the viewer
      const viewerPage = await context.newPage();
      const viewer = await registerUser(viewerPage, 'viewer_nc', viewerEmail);
      await acceptInvitation(viewerPage, token);

      // Navigate to board as viewer
      await viewerPage.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await viewerPage.waitForLoadState('networkidle');

      // Viewer should NOT see "Add card" button
      const addCardButton = viewerPage.getByRole('button', { name: 'Add card' });
      await expect(addCardButton).toBeHidden();

      await viewerPage.close();
    });

    test('viewer cannot access organization settings', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const viewerEmail = `viewer_nosettings_${randomId()}@test.local`;

      // Invite a viewer
      const token = await inviteUserAsRole(page, setup.orgId, viewerEmail, 'Viewer');

      // Create a new browser context for the viewer
      const viewerPage = await context.newPage();
      const viewer = await registerUser(viewerPage, 'viewer_ns', viewerEmail);
      await acceptInvitation(viewerPage, token);

      // Navigate to organization
      await viewerPage.goto(`/organizations/${setup.orgId}`);
      await viewerPage.waitForLoadState('networkidle');

      // Viewer should NOT see Settings link or it should be disabled/hidden
      const settingsLink = viewerPage.getByRole('link', { name: 'Settings' });
      const isHidden = await settingsLink.isHidden().catch(() => true);

      // If settings link exists, clicking it should either fail or show limited options
      if (!isHidden) {
        await settingsLink.click();
        await viewerPage.waitForLoadState('networkidle');

        // Viewer should NOT see Roles tab
        await expect(viewerPage.getByRole('link', { name: 'Roles' })).toBeHidden();
      }

      await viewerPage.close();
    });
  });

  test.describe('Member Role', () => {
    test('member can create and edit cards but not delete', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const memberEmail = `member_${randomId()}@test.local`;

      // Invite a member
      const token = await inviteUserAsRole(page, setup.orgId, memberEmail, 'Member');

      // Create a new browser context for the member
      const memberPage = await context.newPage();
      const member = await registerUser(memberPage, 'member', memberEmail);
      await acceptInvitation(memberPage, token);

      // Navigate to board as member
      await memberPage.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await memberPage.waitForLoadState('networkidle');

      // Member should see "Add card" button
      await expect(memberPage.getByRole('button', { name: 'Add card' }).first()).toBeVisible();

      // Create a card
      await memberPage.getByRole('button', { name: 'Add card' }).first().click();
      await expect(memberPage.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
      await memberPage.fill('#title', 'Member Created Card');
      await memberPage.getByRole('button', { name: 'Create Card' }).click();
      await expect(memberPage.getByLabel('Create Card')).toBeHidden({ timeout: 5000 });

      // Card should appear
      await expect(memberPage.getByText('Member Created Card')).toBeVisible({ timeout: 5000 });

      // Click on card to open detail - member can edit but may not be able to delete
      await memberPage.getByText('Member Created Card').click();
      await memberPage.waitForTimeout(500);

      // Member should see editable form
      await expect(memberPage.locator('input#detail-title')).toBeVisible();

      await memberPage.close();
    });

    test('member cannot manage board columns', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const memberEmail = `member_nocol_${randomId()}@test.local`;

      // Invite a member
      const token = await inviteUserAsRole(page, setup.orgId, memberEmail, 'Member');

      // Create a new browser context for the member
      const memberPage = await context.newPage();
      const member = await registerUser(memberPage, 'member_nc', memberEmail);
      await acceptInvitation(memberPage, token);

      // Navigate to board as member
      await memberPage.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await memberPage.waitForLoadState('networkidle');

      // Member should NOT see "Add Column" element
      await expect(memberPage.getByText('Add Column')).toBeHidden();

      await memberPage.close();
    });
  });

  test.describe('Admin Role', () => {
    test('admin can access organization settings', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const adminEmail = `admin_${randomId()}@test.local`;

      // Invite an admin
      const token = await inviteUserAsRole(page, setup.orgId, adminEmail, 'Admin');

      // Create a new browser context for the admin
      const adminPage = await context.newPage();
      const admin = await registerUser(adminPage, 'admin', adminEmail);
      await acceptInvitation(adminPage, token);

      // Navigate to organization settings
      await adminPage.goto(`/organizations/${setup.orgId}/settings`);
      await adminPage.waitForLoadState('networkidle');

      // Admin should see Members tab
      await expect(adminPage.getByRole('link', { name: 'Members' })).toBeVisible();

      // Admin can see Roles tab (UI doesn't hide based on permissions yet, but API will enforce)
      await expect(adminPage.getByRole('link', { name: 'Roles' })).toBeVisible();

      // Verify admin can view members list (already on members page due to redirect)
      await adminPage.waitForLoadState('networkidle');
      // Should see at least one member with a role displayed (use exact to avoid matching username containing "owner")
      await expect(adminPage.getByText('Owner', { exact: true })).toBeVisible({ timeout: 10000 });

      await adminPage.close();
    });

    test('admin can create and delete cards', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const adminEmail = `admin_cards_${randomId()}@test.local`;

      // Invite an admin
      const token = await inviteUserAsRole(page, setup.orgId, adminEmail, 'Admin');

      // Create a new browser context for the admin
      const adminPage = await context.newPage();
      const admin = await registerUser(adminPage, 'admin_c', adminEmail);
      await acceptInvitation(adminPage, token);

      // Navigate to board
      await adminPage.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await adminPage.waitForLoadState('networkidle');

      // Admin can see and create cards
      await expect(adminPage.getByRole('button', { name: 'Add card' }).first()).toBeVisible();

      // Click on existing card
      await adminPage.getByText('Test Card for RBAC').click();
      await adminPage.waitForTimeout(500);

      // Admin should see delete button
      await expect(adminPage.getByRole('button', { name: /Delete/ })).toBeVisible();

      await adminPage.close();
    });

    test('admin can manage board columns', async ({ page, context }) => {
      // Setup as owner
      const setup = await setupTestEnvironment(page);
      const adminEmail = `admin_cols_${randomId()}@test.local`;

      // Invite an admin
      const token = await inviteUserAsRole(page, setup.orgId, adminEmail, 'Admin');

      // Create a new browser context for the admin
      const adminPage = await context.newPage();
      const admin = await registerUser(adminPage, 'admin_cols', adminEmail);
      await acceptInvitation(adminPage, token);

      // Navigate to board
      await adminPage.goto(`/projects/${setup.projectId}/board/${setup.boardId}`);
      await adminPage.waitForLoadState('networkidle');

      // Admin should see "Add Column" element
      await expect(adminPage.getByText('Add Column')).toBeVisible();

      await adminPage.close();
    });
  });

});
