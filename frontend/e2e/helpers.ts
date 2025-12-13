import { expect, type Page } from '@playwright/test';

// MailHog API URL - connects to Docker mailhog service
const MAILHOG_API_URL = 'http://localhost:8025/api/v2';

/**
 * MailHog message structure
 */
interface MailHogMessage {
  ID: string;
  From: { Mailbox: string; Domain: string };
  To: Array<{ Mailbox: string; Domain: string }>;
  Content: {
    Headers: {
      Subject: string[];
      From: string[];
      To: string[];
    };
    Body: string;
  };
  Raw: {
    From: string;
    To: string[];
    Data: string;
  };
  Created: string;
}

interface MailHogSearchResult {
  total: number;
  count: number;
  start: number;
  items: MailHogMessage[];
}

/**
 * Fetches all emails from MailHog
 */
export async function getMailHogMessages(): Promise<MailHogMessage[]> {
  const response = await fetch(`${MAILHOG_API_URL}/messages`);
  if (!response.ok) {
    throw new Error(`Failed to fetch MailHog messages: ${response.status}`);
  }
  const data: MailHogSearchResult = await response.json();
  return data.items;
}

/**
 * Searches for emails by recipient
 */
export async function searchMailHogByRecipient(email: string): Promise<MailHogMessage[]> {
  const response = await fetch(`${MAILHOG_API_URL}/search?kind=to&query=${encodeURIComponent(email)}`);
  if (!response.ok) {
    throw new Error(`Failed to search MailHog: ${response.status}`);
  }
  const data: MailHogSearchResult = await response.json();
  return data.items;
}

/**
 * Waits for an email to arrive at MailHog for a specific recipient
 */
export async function waitForEmail(email: string, timeout: number = 10000): Promise<MailHogMessage> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    const messages = await searchMailHogByRecipient(email);
    if (messages.length > 0) {
      // Return the most recent message
      return messages[0];
    }
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  throw new Error(`Timeout waiting for email to ${email}`);
}

/**
 * Decodes quoted-printable encoded content
 */
function decodeQuotedPrintable(input: string): string {
  // Replace soft line breaks (=\r\n or =\n)
  let result = input.replace(/=\r?\n/g, '');
  // Decode =XX sequences (like =3D for =)
  result = result.replace(/=([0-9A-Fa-f]{2})/g, (_, hex) => {
    return String.fromCharCode(parseInt(hex, 16));
  });
  return result;
}

/**
 * Extracts the verification token from an email body
 * Looks for the verification URL pattern and extracts the token
 * Handles quoted-printable encoding
 */
export function extractVerificationToken(message: MailHogMessage): string {
  // Decode the quoted-printable encoded body first
  const body = decodeQuotedPrintable(message.Raw.Data);
  // The verification URL pattern: /verify?token=<token>
  const tokenMatch = body.match(/[?&]token=([a-f0-9]+)/i);
  if (!tokenMatch) {
    throw new Error('Could not find verification token in email');
  }
  return tokenMatch[1];
}

/**
 * Deletes all messages from MailHog (useful for test cleanup)
 */
export async function clearMailHog(): Promise<void> {
  await fetch(`${MAILHOG_API_URL.replace('/api/v2', '/api/v1')}/messages`, {
    method: 'DELETE',
  });
}

/**
 * Generates a random ID for unique test data
 */
export function randomId(): string {
  return Math.random().toString(36).substring(2, 10);
}

/**
 * Generates a random uppercase letter string (for project keys)
 */
export function randomLetters(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

export interface TestContext {
  testId: string;
  username: string;
  email: string;
  password: string;
  orgName: string;
  orgId: string;
  projectName: string;
  projectKey: string;
  projectId: string;
  boardId?: string;
}

/**
 * Creates a fresh isolated test environment with a new user, organization, and project.
 * Each test gets its own unique data to ensure complete isolation.
 * By default, email verification is skipped to keep tests fast. Set verifyEmail: true to enable.
 */
export async function setupTestEnvironment(page: Page, prefix: string = 'test', options?: { verifyEmail?: boolean }): Promise<TestContext> {
  const testId = randomId();
  const username = `${prefix}_${testId}`;
  const email = `${username}@test.local`;
  const password = 'testpassword123';
  const orgName = `${prefix} Org ${testId}`;
  const projectName = `${prefix} Project ${testId}`;
  const projectKey = `${prefix.substring(0, 2).toUpperCase()}${randomLetters(4)}`;

  // Register user
  await page.goto('/register');
  await page.waitForLoadState('networkidle');
  await page.fill('#username', username);
  await page.fill('#email', email);
  await page.fill('#password', password);
  await page.fill('#confirmPassword', password);
  const registerButton = page.getByRole('button', { name: 'Register' });
  // Use Promise.all to wait for navigation while clicking
  await Promise.all([
    page.waitForURL('/', { timeout: 20000 }),
    registerButton.click()
  ]);

  // Verify email if requested
  if (options?.verifyEmail) {
    await verifyEmailWithMailHog(page, email);
  }

  // Create organization
  await page.goto('/organizations/new');
  await page.waitForLoadState('networkidle');
  await page.fill('#name', orgName);
  const createOrgButton = page.getByRole('button', { name: 'Create Organization' });
  // Use Promise.all to wait for navigation while clicking
  await Promise.all([
    page.waitForURL(/\/organizations\/([a-f0-9-]+)/, { timeout: 20000 }),
    createOrgButton.click()
  ]);

  const orgUrl = page.url();
  const orgMatch = orgUrl.match(/\/organizations\/([a-f0-9-]+)/);
  const orgId = orgMatch ? orgMatch[1] : '';

  // Create project
  await page.goto(`/organizations/${orgId}/projects/new`);
  await page.waitForLoadState('networkidle');
  await page.fill('#name', projectName);
  await page.fill('#key', projectKey);
  const createProjectButton = page.getByRole('button', { name: 'Create Project' });
  // Use Promise.all to wait for navigation while clicking
  await Promise.all([
    page.waitForURL(/\/projects\/([a-f0-9-]+)/, { timeout: 20000 }),
    createProjectButton.click()
  ]);

  const projectUrl = page.url();
  const projectMatch = projectUrl.match(/\/projects\/([a-f0-9-]+)/);
  const projectId = projectMatch ? projectMatch[1] : '';

  return {
    testId,
    username,
    email,
    password,
    orgName,
    orgId,
    projectName,
    projectKey,
    projectId,
  };
}

/**
 * Verifies a user's email by fetching the verification token from MailHog and visiting the verification URL
 */
export async function verifyEmailWithMailHog(page: Page, email: string): Promise<void> {
  // Wait for the verification email to arrive
  const message = await waitForEmail(email, 15000);

  // Extract the verification token
  const token = extractVerificationToken(message);

  // Visit the verification page
  await page.goto(`/verify?token=${token}`);
  await page.waitForLoadState('networkidle');

  // Wait for verification to complete (should redirect to home or show success)
  // Give it a moment to process
  await page.waitForTimeout(1000);
}

/**
 * Logs in a user
 */
export async function login(page: Page, username: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.fill('#username', username);
  await page.fill('#password', password);
  const signInButton = page.getByRole('button', { name: 'Sign in' });
  // Use Promise.all to wait for navigation while clicking
  await Promise.all([
    page.waitForURL('/', { timeout: 20000 }),
    signInButton.click()
  ]);
}

/**
 * Navigates to the board page for a project.
 * If no boards exist, creates a default "Kanban Board".
 */
export async function navigateToBoard(page: Page, projectId: string): Promise<string> {
  await page.goto(`/projects/${projectId}`);
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

  // Click on the first board link
  await page.locator('a[href*="/board/"]').first().click();
  await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+\/board\/[a-f0-9-]+/, { timeout: 10000 });
  await expect(page.getByRole('heading', { name: 'Todo', exact: true })).toBeVisible({ timeout: 10000 });

  const boardUrl = page.url();
  const boardMatch = boardUrl.match(/\/board\/([a-f0-9-]+)/);
  return boardMatch ? boardMatch[1] : '';
}

/**
 * Gets a column locator by name
 */
export function getColumn(page: Page, columnName: string) {
  return page.locator('.w-72').filter({ has: page.locator(`h3:has-text("${columnName}")`) });
}

/**
 * Clicks add card button in a specific column
 */
export async function clickAddCardInColumn(page: Page, columnName: string) {
  await getColumn(page, columnName).getByRole('button', { name: 'Add card' }).click();
  await expect(page.getByRole('heading', { name: 'Create Card' })).toBeVisible({ timeout: 5000 });
}

/**
 * Fills a TipTap rich text editor with text.
 * The editor uses a .ProseMirror contenteditable element.
 * @param page - Playwright page
 * @param containerSelector - Selector for the container that holds the RichTextEditor (optional, defaults to first .ProseMirror)
 * @param text - Text to enter
 */
export async function fillRichTextEditor(page: Page, text: string, containerSelector?: string) {
  // Find the ProseMirror editor element
  const editorLocator = containerSelector
    ? page.locator(containerSelector).locator('.ProseMirror')
    : page.locator('.ProseMirror').first();

  // Click to focus
  await editorLocator.click();

  // Select all and delete (to clear any existing content)
  await page.keyboard.press('Meta+a');
  await page.keyboard.press('Backspace');

  // Type the new text
  await page.keyboard.type(text);
}

/**
 * Creates a card in a specific column and waits for it to appear
 */
export async function createCard(page: Page, columnName: string, title: string, description?: string) {
  await clickAddCardInColumn(page, columnName);
  await page.fill('#title', title);
  if (description) {
    await fillRichTextEditor(page, description);
  }
  await page.getByRole('button', { name: 'Create Card' }).click();
  await expect(page.getByLabel('Create Card')).toBeHidden({ timeout: 5000 });
  await expect(page.getByText(title)).toBeVisible({ timeout: 5000 });
}
