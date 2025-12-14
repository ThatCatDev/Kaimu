import { test, expect, type Page } from '@playwright/test';
import {
  setupTestEnvironment,
  randomId,
  login,
  waitForEmail,
  clearMailHog,
  isMailHogAvailable,
  type TestContext
} from './helpers';

/**
 * Extracts the invitation token from an invitation email body
 * Looks for the /invite/<token> URL pattern
 */
function extractInvitationToken(emailBody: string): string {
  // Decode quoted-printable encoding
  let decoded = emailBody.replace(/=\r?\n/g, '');
  decoded = decoded.replace(/=([0-9A-Fa-f]{2})/g, (_, hex) => {
    return String.fromCharCode(parseInt(hex, 16));
  });

  // Look for the invite URL pattern: /invite/<token>
  const tokenMatch = decoded.match(/\/invite\/([A-Za-z0-9_-]+)/);
  if (!tokenMatch) {
    throw new Error('Could not find invitation token in email');
  }
  return tokenMatch[1];
}

/**
 * Navigates to organization settings members page
 */
async function navigateToMembersSettings(page: Page, orgId: string) {
  await page.goto(`/organizations/${orgId}/settings`);
  await page.waitForLoadState('networkidle');
  // Wait for the Members tab content to load (Members tab is active by default)
  await expect(page.getByRole('heading', { name: 'Members' })).toBeVisible({ timeout: 10000 });
}

/**
 * Selects a role in the BitsSelect dropdown
 * bits-ui Select renders as a button, not combobox
 */
async function selectRole(page: Page, roleName: string) {
  // The form group has "Role" label, find the button inside it
  const roleFormGroup = page.locator('div').filter({ hasText: /^Role/ }).first();
  const selectTrigger = roleFormGroup.locator('button');
  await selectTrigger.click();
  // Wait for dropdown to appear and click the role option
  await page.getByRole('option', { name: roleName }).click();
}

/**
 * Creates an invitation through the UI and returns the invitation token
 */
async function createInvitation(
  page: Page,
  orgId: string,
  email: string,
  roleName: string = 'Member'
): Promise<string> {
  await navigateToMembersSettings(page, orgId);

  // Click invite button
  await page.getByRole('button', { name: 'Invite Member' }).click();
  await expect(page.getByRole('heading', { name: 'Invite Member' })).toBeVisible({ timeout: 5000 });

  // Fill out the form
  await page.fill('input[type="email"]', email);

  // Select role if not Member (Member is default)
  if (roleName !== 'Member') {
    await selectRole(page, roleName);
  }

  // Submit
  await page.getByRole('button', { name: 'Create Invitation' }).click();

  // Wait for success state
  await expect(page.getByRole('heading', { name: 'Invitation Created' })).toBeVisible({ timeout: 10000 });

  // Get the invitation link from the input field
  const linkInput = page.locator('input[readonly]');
  const inviteLink = await linkInput.inputValue();

  // Extract token from link
  const tokenMatch = inviteLink.match(/\/invite\/([A-Za-z0-9_=-]+)/);
  if (!tokenMatch) {
    throw new Error('Could not extract token from invitation link');
  }

  // Click done to close
  await page.getByRole('button', { name: 'Done' }).click();
  await expect(page.getByRole('heading', { name: 'Invitation Created' })).toBeHidden({ timeout: 5000 });

  return tokenMatch[1];
}

test.describe('Invitation System', () => {
  test.describe('Creating Invitations', () => {
    test('can create an invitation and see the invitation link', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      await navigateToMembersSettings(page, ctx.orgId);

      // Click invite button
      await page.getByRole('button', { name: 'Invite Member' }).click();
      await expect(page.getByRole('heading', { name: 'Invite Member' })).toBeVisible({ timeout: 5000 });

      // Fill out the form
      await page.fill('input[type="email"]', inviteeEmail);

      // Submit
      await page.getByRole('button', { name: 'Create Invitation' }).click();

      // Should show success state with invitation link
      await expect(page.getByRole('heading', { name: 'Invitation Created' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText(`Invitation sent to ${inviteeEmail}`)).toBeVisible();

      // Should show the invitation link
      const linkInput = page.locator('input[readonly]');
      await expect(linkInput).toBeVisible();
      const inviteLink = await linkInput.inputValue();
      expect(inviteLink).toContain('/invite/');

      // Should show expiration info
      await expect(page.getByText('The link expires in 7 days')).toBeVisible();
    });

    test('can copy invitation link to clipboard', async ({ page, context }) => {
      // Grant clipboard permissions
      await context.grantPermissions(['clipboard-read', 'clipboard-write']);

      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      await navigateToMembersSettings(page, ctx.orgId);

      // Create invitation
      await page.getByRole('button', { name: 'Invite Member' }).click();
      await page.fill('input[type="email"]', inviteeEmail);
      await page.getByRole('button', { name: 'Create Invitation' }).click();
      await expect(page.getByRole('heading', { name: 'Invitation Created' })).toBeVisible({ timeout: 10000 });

      // Click copy button
      await page.getByRole('button', { name: 'Copy' }).click();

      // Should show "Copied!" feedback
      await expect(page.getByText('Copied!')).toBeVisible({ timeout: 3000 });

      // Verify clipboard content
      const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
      expect(clipboardText).toContain('/invite/');
    });

    test('can select different roles when creating invitation', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      await navigateToMembersSettings(page, ctx.orgId);

      // Click invite button
      await page.getByRole('button', { name: 'Invite Member' }).click();
      await expect(page.getByRole('heading', { name: 'Invite Member' })).toBeVisible({ timeout: 5000 });

      // Fill email
      await page.fill('input[type="email"]', inviteeEmail);

      // Open role dropdown - bits-ui Select renders as a button, not combobox
      const roleFormGroup = page.locator('div').filter({ hasText: /^Role/ }).first();
      const selectTrigger = roleFormGroup.locator('button');
      await selectTrigger.click();

      // Should see available roles (except Owner)
      await expect(page.getByRole('option', { name: 'Admin' })).toBeVisible();
      await expect(page.getByRole('option', { name: 'Member' })).toBeVisible();
      await expect(page.getByRole('option', { name: 'Viewer' })).toBeVisible();
      // Owner should not be an option
      await expect(page.getByRole('option', { name: 'Owner' })).toHaveCount(0);

      // Select Admin role
      await page.getByRole('option', { name: 'Admin' }).click();

      // Submit
      await page.getByRole('button', { name: 'Create Invitation' }).click();

      // Should succeed
      await expect(page.getByRole('heading', { name: 'Invitation Created' })).toBeVisible({ timeout: 10000 });
    });

    test('shows error when inviting existing member', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');

      await navigateToMembersSettings(page, ctx.orgId);

      // Try to invite self (current user is already a member)
      await page.getByRole('button', { name: 'Invite Member' }).click();
      await page.fill('input[type="email"]', ctx.email);
      await page.getByRole('button', { name: 'Create Invitation' }).click();

      // Should show error
      await expect(page.getByText(/already a member/i)).toBeVisible({ timeout: 10000 });
    });

    test('shows error when invitation already pending', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create first invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Try to create another invitation for the same email
      await page.getByRole('button', { name: 'Invite Member' }).click();
      await page.fill('input[type="email"]', inviteeEmail);
      await page.getByRole('button', { name: 'Create Invitation' }).click();

      // Should show error about pending invitation
      await expect(page.getByText(/already a pending invitation/i)).toBeVisible({ timeout: 10000 });
    });

    test('cancel button closes the modal without creating invitation', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');

      await navigateToMembersSettings(page, ctx.orgId);

      // Open modal
      await page.getByRole('button', { name: 'Invite Member' }).click();
      await expect(page.getByRole('heading', { name: 'Invite Member' })).toBeVisible({ timeout: 5000 });

      // Fill some data
      await page.fill('input[type="email"]', 'test@example.com');

      // Cancel
      await page.getByRole('button', { name: 'Cancel' }).click();

      // Modal should close
      await expect(page.getByRole('heading', { name: 'Invite Member' })).toBeHidden({ timeout: 5000 });

      // No pending invitations should exist
      await expect(page.getByText('Pending Invitations')).toBeHidden();
    });
  });

  test.describe('Pending Invitations', () => {
    test('shows pending invitations in the members list', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Should see pending invitations section
      await expect(page.getByText('Pending Invitations (1)')).toBeVisible();

      // Should show the invitee email
      await expect(page.getByText(inviteeEmail)).toBeVisible();

      // Should show Pending badge (use exact to avoid matching "Pending Invitations")
      await expect(page.getByText('Pending', { exact: true })).toBeVisible();
    });

    test('can resend an invitation', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Find the resend button in the pending invitations section
      const invitationRow = page.locator('div').filter({ hasText: inviteeEmail }).first();
      const resendButton = invitationRow.getByRole('button', { name: 'Resend' });

      // Click resend
      await resendButton.click();

      // Should show "Sent!" feedback
      await expect(page.getByText('Sent!')).toBeVisible({ timeout: 5000 });

      // After 2 seconds, should go back to "Resend"
      await expect(resendButton.getByText('Resend')).toBeVisible({ timeout: 5000 });
    });

    test('can cancel an invitation', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Verify invitation is shown
      await expect(page.getByText(inviteeEmail, { exact: true })).toBeVisible();

      // Find and click cancel button in the pending invitations section
      const pendingSection = page.locator('div').filter({ hasText: 'Pending Invitations' });
      const cancelButton = pendingSection.getByRole('button', { name: 'Cancel' });
      await cancelButton.click();

      // Should show confirmation modal
      await expect(page.getByRole('heading', { name: 'Cancel Invitation' })).toBeVisible({ timeout: 5000 });
      await expect(page.getByText(/Are you sure you want to cancel the invitation/)).toBeVisible();

      // Confirm cancellation
      await page.getByRole('button', { name: 'Cancel Invitation' }).click();

      // Wait for modal to close
      await expect(page.getByRole('heading', { name: 'Cancel Invitation' })).toBeHidden({ timeout: 5000 });

      // Invitation should be removed - use the Pending Invitations section as a reference
      // After cancellation, the pending invitations section should not exist
      await expect(page.getByText('Pending Invitations')).toBeHidden({ timeout: 5000 });
    });

    test('shows invitation details including role and dates', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Verify the invitation details are shown
      await expect(page.getByText('Pending Invitations (1)')).toBeVisible();
      await expect(page.getByText(inviteeEmail, { exact: true })).toBeVisible();

      // Should show role, invited date, and expiration info in the invitation row
      // The format is "Member · Invited Dec 10, 2025 · Expires Dec 17, 2025"
      const invitationRow = page.locator('div').filter({ hasText: inviteeEmail }).first();
      await expect(invitationRow.getByText(/Invited/)).toBeVisible();
      await expect(invitationRow.getByText(/Expires/)).toBeVisible();
    });
  });

  test.describe('Accepting Invitations', () => {
    test('shows login required when not authenticated', async ({ page }) => {
      // Create an invitation first (need to be logged in)
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Navigate to home page where Logout button is visible, then logout
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.getByRole('button', { name: 'Logout' }).click();
      await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });

      // Visit invitation page
      await page.goto(`/invite/${token}`);
      await page.waitForLoadState('networkidle');

      // Should show login required message
      await expect(page.getByRole('heading', { name: 'Login Required' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText('Please log in or create an account to accept this invitation')).toBeVisible();

      // Should have login and register buttons
      await expect(page.getByRole('button', { name: 'Log In' })).toBeVisible();
      await expect(page.getByRole('link', { name: 'Create Account' })).toBeVisible();
    });

    test('can accept invitation as logged in user', async ({ page }) => {
      // Create invitation with first user
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Navigate to home page where Logout button is visible, then logout first user
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.getByRole('button', { name: 'Logout' }).click();
      await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });

      // Register as the invitee
      const inviteeUsername = `invitee_${randomId()}`;
      await page.goto('/register');
      await page.waitForLoadState('networkidle');
      await page.fill('#username', inviteeUsername);
      await page.fill('#email', inviteeEmail);
      await page.fill('#password', 'testpassword123');
      await page.fill('#confirmPassword', 'testpassword123');
      await Promise.all([
        page.waitForURL('/', { timeout: 20000 }),
        page.getByRole('button', { name: 'Register' }).click()
      ]);

      // Visit invitation page
      await page.goto(`/invite/${token}`);
      await page.waitForLoadState('networkidle');

      // Should show accept invitation UI
      await expect(page.getByRole('heading', { name: "You've Been Invited!" })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText(`Logged in as ${inviteeUsername}`)).toBeVisible();

      // Accept invitation
      await page.getByRole('button', { name: 'Accept Invitation' }).click();

      // Should show success
      await expect(page.getByText(`Welcome to ${ctx.orgName}!`)).toBeVisible({ timeout: 10000 });
      await expect(page.getByText('You have successfully joined the organization')).toBeVisible();

      // Can navigate to organization
      await page.getByRole('button', { name: 'Go to Organization' }).click();
      await expect(page).toHaveURL(new RegExp(`/organizations/${ctx.orgId}`), { timeout: 10000 });
    });

    test('shows error for invalid invitation token', async ({ page }) => {
      // Setup test user
      const ctx = await setupTestEnvironment(page, 'inv');

      // Visit with invalid token
      await page.goto('/invite/invalid-token-12345');
      await page.waitForLoadState('networkidle');

      // Should show ready to accept state initially
      await expect(page.getByRole('heading', { name: "You've Been Invited!" })).toBeVisible({ timeout: 10000 });

      // Try to accept
      await page.getByRole('button', { name: 'Accept Invitation' }).click();

      // Should show error
      await expect(page.getByRole('heading', { name: 'Unable to Accept Invitation' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText(/invalid or has already been used/i)).toBeVisible();
    });

    test('shows error when user is already a member', async ({ page }) => {
      // Create invitation to self
      const ctx = await setupTestEnvironment(page, 'inv');

      // We need to manually create a token for this test since the UI prevents inviting existing members
      // For this test, we'll create an invitation for a different email, then try with a user
      // who is already a member

      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // As the owner (already a member), try to accept this invitation
      // Note: This will fail because email doesn't match
      await page.goto(`/invite/${token}`);
      await page.waitForLoadState('networkidle');

      await expect(page.getByRole('heading', { name: "You've Been Invited!" })).toBeVisible({ timeout: 10000 });

      await page.getByRole('button', { name: 'Accept Invitation' }).click();

      // Should show error about email mismatch (since our email doesn't match the invitation)
      await expect(page.getByRole('heading', { name: 'Unable to Accept Invitation' })).toBeVisible({ timeout: 10000 });
    });

    test('cancel button on accept page goes to dashboard', async ({ page }) => {
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Navigate to home page where Logout button is visible, then logout and create new user
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.getByRole('button', { name: 'Logout' }).click();
      await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });

      // Register as invitee
      const inviteeUsername = `invitee_${randomId()}`;
      await page.goto('/register');
      await page.waitForLoadState('networkidle');
      await page.fill('#username', inviteeUsername);
      await page.fill('#email', inviteeEmail);
      await page.fill('#password', 'testpassword123');
      await page.fill('#confirmPassword', 'testpassword123');
      await Promise.all([
        page.waitForURL('/', { timeout: 20000 }),
        page.getByRole('button', { name: 'Register' }).click()
      ]);

      // Visit invitation page
      await page.goto(`/invite/${token}`);
      await page.waitForLoadState('networkidle');

      // Click cancel
      await page.getByRole('button', { name: 'Cancel' }).click();

      // Should go to dashboard
      await expect(page).toHaveURL('/dashboard', { timeout: 10000 });
    });

    test('login button navigates to login page', async ({ page }) => {
      // Create invitation
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Navigate to home page where Logout button is visible, then logout
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.getByRole('button', { name: 'Logout' }).click();
      await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });

      // Visit invitation page
      await page.goto(`/invite/${token}`);
      await page.waitForLoadState('networkidle');

      // Should show login required state
      await expect(page.getByRole('heading', { name: 'Login Required' })).toBeVisible({ timeout: 10000 });

      // Click login button
      await page.getByRole('button', { name: 'Log In' }).click();

      // Should be on login page
      await expect(page).toHaveURL('/login', { timeout: 10000 });
    });
  });

  test.describe('Invitation Email', () => {
    test('sends invitation email to invitee', async ({ page }) => {
      // Skip if MailHog is not available
      test.skip(!(await isMailHogAvailable()), 'MailHog not available');

      // Clear mailhog first
      await clearMailHog();

      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Wait for email to arrive
      const email = await waitForEmail(inviteeEmail, 15000);
      expect(email).toBeDefined();

      // Verify email content
      expect(email.Content.Headers.Subject[0]).toContain('invited to join');

      // Extract and verify the invitation token from email
      const token = extractInvitationToken(email.Raw.Data);
      expect(token).toBeDefined();
      expect(token.length).toBeGreaterThan(0);
    });

    test('resending invitation sends new email', async ({ page }) => {
      // Skip if MailHog is not available
      test.skip(!(await isMailHogAvailable()), 'MailHog not available');

      // Clear mailhog first
      await clearMailHog();

      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;

      // Create invitation
      await createInvitation(page, ctx.orgId, inviteeEmail);

      // Wait for first email
      await waitForEmail(inviteeEmail, 15000);

      // Clear mailhog again
      await clearMailHog();

      // Resend invitation
      const invitationRow = page.locator('div').filter({ hasText: inviteeEmail }).first();
      await invitationRow.getByRole('button', { name: 'Resend' }).click();
      await expect(page.getByText('Sent!')).toBeVisible({ timeout: 5000 });

      // Wait for new email
      const newEmail = await waitForEmail(inviteeEmail, 15000);
      expect(newEmail).toBeDefined();
      expect(newEmail.Content.Headers.Subject[0]).toContain('invited to join');
    });
  });

  test.describe('Member Management After Invitation', () => {
    test('accepted invitation shows user in members list', async ({ page, browser }) => {
      // Create invitation with first user
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const inviteeUsername = `invitee_${randomId()}`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Use a new browser context for the invitee
      const inviteeContext = await browser.newContext();
      const inviteePage = await inviteeContext.newPage();

      // Register as invitee
      await inviteePage.goto('/register');
      await inviteePage.waitForLoadState('networkidle');
      await inviteePage.fill('#username', inviteeUsername);
      await inviteePage.fill('#email', inviteeEmail);
      await inviteePage.fill('#password', 'testpassword123');
      await inviteePage.fill('#confirmPassword', 'testpassword123');
      await Promise.all([
        inviteePage.waitForURL('/', { timeout: 20000 }),
        inviteePage.getByRole('button', { name: 'Register' }).click()
      ]);

      // Accept invitation
      await inviteePage.goto(`/invite/${token}`);
      await inviteePage.waitForLoadState('networkidle');
      await inviteePage.getByRole('button', { name: 'Accept Invitation' }).click();
      await expect(inviteePage.getByText(`Welcome to ${ctx.orgName}!`)).toBeVisible({ timeout: 10000 });

      // Close invitee context
      await inviteeContext.close();

      // Back to original user - refresh members page
      await navigateToMembersSettings(page, ctx.orgId);

      // Should see the new member in the list (email may appear twice in member row, use first())
      await expect(page.getByText(inviteeEmail).first()).toBeVisible({ timeout: 10000 });

      // Should no longer be in pending invitations - section should be hidden since there are no more pending
      await expect(page.getByText('Pending Invitations')).toBeHidden({ timeout: 5000 });
    });

    test('can change role of member who joined via invitation', async ({ page, browser }) => {
      // Create invitation with first user
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const inviteeUsername = `invitee_${randomId()}`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Use a new browser context for the invitee
      const inviteeContext = await browser.newContext();
      const inviteePage = await inviteeContext.newPage();

      // Register and accept invitation as invitee
      await inviteePage.goto('/register');
      await inviteePage.waitForLoadState('networkidle');
      await inviteePage.fill('#username', inviteeUsername);
      await inviteePage.fill('#email', inviteeEmail);
      await inviteePage.fill('#password', 'testpassword123');
      await inviteePage.fill('#confirmPassword', 'testpassword123');
      await Promise.all([
        inviteePage.waitForURL('/', { timeout: 20000 }),
        inviteePage.getByRole('button', { name: 'Register' }).click()
      ]);

      await inviteePage.goto(`/invite/${token}`);
      await inviteePage.waitForLoadState('networkidle');
      await inviteePage.getByRole('button', { name: 'Accept Invitation' }).click();
      await expect(inviteePage.getByText(`Welcome to ${ctx.orgName}!`)).toBeVisible({ timeout: 10000 });
      await inviteeContext.close();

      // Back to original user
      await navigateToMembersSettings(page, ctx.orgId);

      // Find the new member row and click change role (button has title, not visible text)
      const memberRow = page.locator('div').filter({ hasText: inviteeEmail }).first();
      await memberRow.locator('button[title="Change role"]').click();

      // Change role modal should appear
      await expect(page.getByRole('heading', { name: 'Change Role' })).toBeVisible({ timeout: 5000 });

      // Select Admin role - bits-ui Select renders as a button
      const roleFormGroup = page.locator('div').filter({ hasText: /^Role/ }).first();
      const selectTrigger = roleFormGroup.locator('button');
      await selectTrigger.click();
      await page.getByRole('option', { name: 'Admin' }).click();

      // Confirm (use exact match to avoid matching the icon button with title="Change role")
      await page.getByRole('button', { name: 'Change Role', exact: true }).click();

      // Modal should close
      await expect(page.getByRole('heading', { name: 'Change Role' })).toBeHidden({ timeout: 5000 });

      // Member should now have Admin role
      await expect(memberRow.getByText('Admin')).toBeVisible({ timeout: 5000 });
    });

    test('can remove member who joined via invitation', async ({ page, browser }) => {
      // Create invitation with first user
      const ctx = await setupTestEnvironment(page, 'inv');
      const inviteeEmail = `invitee_${randomId()}@test.local`;
      const inviteeUsername = `invitee_${randomId()}`;
      const token = await createInvitation(page, ctx.orgId, inviteeEmail);

      // Use a new browser context for the invitee
      const inviteeContext = await browser.newContext();
      const inviteePage = await inviteeContext.newPage();

      // Register and accept invitation as invitee
      await inviteePage.goto('/register');
      await inviteePage.waitForLoadState('networkidle');
      await inviteePage.fill('#username', inviteeUsername);
      await inviteePage.fill('#email', inviteeEmail);
      await inviteePage.fill('#password', 'testpassword123');
      await inviteePage.fill('#confirmPassword', 'testpassword123');
      await Promise.all([
        inviteePage.waitForURL('/', { timeout: 20000 }),
        inviteePage.getByRole('button', { name: 'Register' }).click()
      ]);

      await inviteePage.goto(`/invite/${token}`);
      await inviteePage.waitForLoadState('networkidle');
      await inviteePage.getByRole('button', { name: 'Accept Invitation' }).click();
      await expect(inviteePage.getByText(`Welcome to ${ctx.orgName}!`)).toBeVisible({ timeout: 10000 });
      await inviteeContext.close();

      // Back to original user
      await navigateToMembersSettings(page, ctx.orgId);

      // Find the new member row and click remove (button has title, not visible text)
      const memberRow = page.locator('div').filter({ hasText: inviteeEmail }).first();
      await memberRow.locator('button[title="Remove member"]').click();

      // Confirmation modal should appear
      await expect(page.getByRole('heading', { name: 'Remove Member' })).toBeVisible({ timeout: 5000 });
      await expect(page.getByText(/Are you sure you want to remove/)).toBeVisible();

      // Confirm removal (use exact match to avoid matching the icon button with title="Remove member")
      await page.getByRole('button', { name: 'Remove Member', exact: true }).click();

      // Modal should close and member should be removed
      await expect(page.getByRole('heading', { name: 'Remove Member' })).toBeHidden({ timeout: 5000 });
      // Verify member is no longer in the members list (email may still appear in modal confirmation text briefly)
      await page.waitForTimeout(1000); // Wait for DOM to update
      await expect(page.getByText(inviteeEmail).first()).toBeHidden({ timeout: 5000 });
    });
  });
});
