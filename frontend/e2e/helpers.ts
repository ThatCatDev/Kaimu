import { expect, type Page } from '@playwright/test';

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
 */
export async function setupTestEnvironment(page: Page, prefix: string = 'test'): Promise<TestContext> {
  const testId = randomId();
  const username = `${prefix}_${testId}`;
  const password = 'testpassword123';
  const orgName = `${prefix} Org ${testId}`;
  const projectName = `${prefix} Project ${testId}`;
  const projectKey = `${prefix.substring(0, 2).toUpperCase()}${randomLetters(4)}`;

  // Register user
  await page.goto('/register');
  await page.waitForLoadState('networkidle');
  await page.fill('#username', username);
  await page.fill('#password', password);
  await page.fill('#confirmPassword', password);
  const registerButton = page.getByRole('button', { name: 'Register' });
  // Use Promise.all to wait for navigation while clicking
  await Promise.all([
    page.waitForURL('/', { timeout: 20000 }),
    registerButton.click()
  ]);

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
    password,
    orgName,
    orgId,
    projectName,
    projectKey,
    projectId,
  };
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
 * Navigates to the board page for a project
 */
export async function navigateToBoard(page: Page, projectId: string): Promise<string> {
  await page.goto(`/projects/${projectId}`);
  await page.waitForLoadState('networkidle');
  await page.getByRole('link', { name: /Kanban Board/ }).click();
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
 * Creates a card in a specific column and waits for it to appear
 */
export async function createCard(page: Page, columnName: string, title: string, description?: string) {
  await clickAddCardInColumn(page, columnName);
  await page.fill('#title', title);
  if (description) {
    await page.fill('#description', description);
  }
  await page.getByRole('button', { name: 'Create Card' }).click();
  await expect(page.getByLabel('Create Card')).toBeHidden({ timeout: 5000 });
  await expect(page.getByText(title)).toBeVisible({ timeout: 5000 });
}
