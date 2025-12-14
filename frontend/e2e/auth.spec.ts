import { test, expect } from '@playwright/test';
import { randomId, waitForEmail, extractVerificationToken, clearMailHog, isMailHogAvailable } from './helpers';

test.describe('Authentication', () => {
  test('homepage shows login and register links when not authenticated', async ({ page }) => {
    // Clear cookies to ensure logged out state
    await page.context().clearCookies();
    await page.goto('/');

    // Wait for page to fully load and hydrate
    await page.waitForLoadState('networkidle');

    // Wait for nav to hydrate - either Login link appears or Loading disappears
    await expect(page.locator('nav').getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 15000 });
    await expect(page.locator('nav').getByRole('link', { name: 'Register' })).toBeVisible();
  });

  test('can navigate to register page', async ({ page }) => {
    await page.goto('/');

    // Wait for nav to hydrate
    await expect(page.getByText('Loading...')).toBeHidden({ timeout: 10000 });

    // Click the Register link in the nav
    await page.locator('nav').getByRole('link', { name: 'Register' }).click();

    await expect(page).toHaveURL('/register');
    await expect(page.getByRole('heading', { name: 'Create your account' })).toBeVisible();
  });

  test('can navigate to login page', async ({ page }) => {
    await page.goto('/');

    // Wait for nav to hydrate
    await expect(page.getByText('Loading...')).toBeHidden({ timeout: 10000 });

    // Click the Login link in the nav
    await page.locator('nav').getByRole('link', { name: 'Login' }).click();

    await expect(page).toHaveURL('/login');
    await expect(page.getByRole('heading', { name: 'Sign in to your account' })).toBeVisible();
  });

  test('register form shows validation error when fields are empty', async ({ page }) => {
    await page.goto('/register');
    // Wait for hydration
    await page.waitForTimeout(500);

    await page.getByRole('button', { name: 'Register' }).click();

    // HTML5 validation should prevent submission
    await expect(page.locator('#username')).toBeFocused();
  });

  test('register form shows error when passwords do not match', async ({ page }) => {
    await page.goto('/register');
    // Wait for hydration
    await page.waitForTimeout(500);

    await page.fill('#username', 'testuser');
    await page.fill('#email', 'testuser@test.local');
    await page.fill('#password', 'password123');
    await page.fill('#confirmPassword', 'differentpassword');
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page.getByText('Passwords do not match')).toBeVisible();
  });

  test('can register a new user', async ({ page }) => {
    const uniqueUser = `e2e_${randomId()}`;
    const email = `${uniqueUser}@test.local`;
    const password = 'testpassword123';

    await page.goto('/register');
    // Wait for hydration
    await page.waitForTimeout(500);

    await page.fill('#username', uniqueUser);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();

    // Should redirect to home and show username in nav
    await expect(page).toHaveURL('/', { timeout: 10000 });
    await expect(page.getByText(`Hello, ${uniqueUser}`)).toBeVisible({ timeout: 10000 });
  });

  test('shows error when registering with existing username', async ({ page }) => {
    const uniqueUser = `e2e_${randomId()}`;
    const email = `${uniqueUser}@test.local`;
    const password = 'testpassword123';

    // First, register the user
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Now try to register again with the same username
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#email', `${uniqueUser}2@test.local`);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page.getByText('username already taken')).toBeVisible({ timeout: 10000 });
  });

  test('can login with registered user', async ({ page }) => {
    const uniqueUser = `e2e_${randomId()}`;
    const email = `${uniqueUser}@test.local`;
    const password = 'testpassword123';

    // First, register the user
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Logout
    await page.getByRole('button', { name: 'Logout' }).click();
    await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });

    // Now login with the registered user
    await page.goto('/login');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#password', password);
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Should redirect to home and show username in nav
    await expect(page).toHaveURL('/', { timeout: 10000 });
    await expect(page.getByText(`Hello, ${uniqueUser}`)).toBeVisible({ timeout: 10000 });
  });

  test('shows error when login with wrong password', async ({ page }) => {
    const uniqueUser = `e2e_${randomId()}`;
    const email = `${uniqueUser}@test.local`;
    const password = 'testpassword123';

    // First, register the user
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Logout
    await page.getByRole('button', { name: 'Logout' }).click();
    await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });

    // Now try to login with wrong password
    await page.goto('/login');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#password', 'wrongpassword');
    await page.getByRole('button', { name: 'Sign in' }).click();

    await expect(page.getByText('invalid username or password')).toBeVisible({ timeout: 10000 });
  });

  test('can logout after login', async ({ page }) => {
    const uniqueUser = `e2e_${randomId()}`;
    const email = `${uniqueUser}@test.local`;
    const password = 'testpassword123';

    // Register and login
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });
    await expect(page.getByText(`Hello, ${uniqueUser}`)).toBeVisible({ timeout: 10000 });

    // Then logout
    await page.getByRole('button', { name: 'Logout' }).click();

    // Should show login/register links again
    await expect(page.getByRole('link', { name: 'Login' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: 'Register' })).toBeVisible();
  });

  test('login page has link to register', async ({ page }) => {
    await page.goto('/login');

    await expect(page.getByText("Don't have an account?")).toBeVisible();
    await page.getByRole('link', { name: 'Register' }).click();

    await expect(page).toHaveURL('/register');
  });

  test('register page has link to login', async ({ page }) => {
    await page.goto('/register');

    await expect(page.getByText('Already have an account?')).toBeVisible();
    await page.getByRole('link', { name: 'Sign in' }).click();

    await expect(page).toHaveURL('/login');
  });

  test('can verify email after registration', async ({ page }) => {
    // Skip if MailHog is not available
    test.skip(!(await isMailHogAvailable()), 'MailHog not available');

    const uniqueUser = `e2e_${randomId()}`;
    const email = `${uniqueUser}@test.local`;
    const password = 'testpassword123';

    // Clear any existing emails in MailHog
    await clearMailHog();

    // Register user
    await page.goto('/register');
    await page.waitForTimeout(500);
    await page.fill('#username', uniqueUser);
    await page.fill('#email', email);
    await page.fill('#password', password);
    await page.fill('#confirmPassword', password);
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Wait for verification email to arrive
    const emailMessage = await waitForEmail(email, 15000);
    expect(emailMessage).toBeDefined();

    // Extract verification token
    const token = extractVerificationToken(emailMessage);
    expect(token).toBeDefined();
    expect(token.length).toBeGreaterThan(0);

    // Visit verification page
    await page.goto(`/verify?token=${token}`);
    await page.waitForLoadState('networkidle');

    // Should show success message
    await expect(page.getByText('Email Verified!')).toBeVisible({ timeout: 10000 });

    // Should redirect to home
    await expect(page).toHaveURL('/', { timeout: 5000 });
  });
});
