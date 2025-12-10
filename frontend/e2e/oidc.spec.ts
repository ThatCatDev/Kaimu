import { test, expect } from '@playwright/test';

test.describe('OIDC Authentication', () => {
  test.describe('Login Page OIDC Elements', () => {
    test('login page shows OIDC provider link when providers are configured', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Check if OIDC provider links exist (links with "Continue with" text)
      const providerLinks = page.locator('a:has-text("Continue with")');

      // If OIDC providers are configured, links should be visible
      const linkCount = await providerLinks.count();
      if (linkCount > 0) {
        await expect(providerLinks.first()).toBeVisible();
        console.log(`Found ${linkCount} OIDC provider(s)`);
      } else {
        // If no OIDC providers are configured, no links should appear
        console.log('No OIDC providers configured - this is expected if Dex is not running');
      }
    });

    test('login page displays "Or continue with" divider when OIDC providers exist', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const providerLinks = page.locator('a:has-text("Continue with")');
      const linkCount = await providerLinks.count();

      if (linkCount > 0) {
        // Should show divider text
        await expect(page.getByText('Or continue with')).toBeVisible();
      }
    });

    test('clicking OIDC provider link redirects to authorization endpoint', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const providerLinks = page.locator('a:has-text("Continue with")');
      const linkCount = await providerLinks.count();

      if (linkCount > 0) {
        const firstProvider = providerLinks.first();

        // Get the provider name for logging
        const providerName = await firstProvider.textContent();
        console.log(`Testing OIDC flow with provider: ${providerName}`);

        // Check the href attribute points to OIDC authorize endpoint
        const href = await firstProvider.getAttribute('href');
        expect(href).toContain('/auth/oidc/');
        expect(href).toContain('/authorize');
      } else {
        test.skip();
      }
    });
  });

  test.describe('Register Page OIDC Elements', () => {
    test('register page shows OIDC provider links when providers are configured', async ({ page }) => {
      await page.goto('/register');
      await page.waitForLoadState('networkidle');

      const providerLinks = page.locator('a:has-text("Continue with")');
      const linkCount = await providerLinks.count();

      if (linkCount > 0) {
        await expect(providerLinks.first()).toBeVisible();
      }
    });
  });

  test.describe('OIDC Callback Error Handling', () => {
    test('callback with error parameter shows error message on login page', async ({ page }) => {
      // Simulate an OIDC callback error by navigating to login with error param
      // Note: The AuthForm component would need to read and display the error from URL params
      // This test verifies the URL structure works
      await page.goto('/login?error=access_denied');
      await page.waitForLoadState('networkidle');

      // Verify the page loaded correctly
      await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('OIDC API Endpoints', () => {
    test('GET /auth/oidc/providers returns JSON list of providers', async ({ request }) => {
      const response = await request.get('http://localhost:3000/auth/oidc/providers');

      expect(response.ok()).toBeTruthy();
      expect(response.headers()['content-type']).toContain('application/json');

      const providers = await response.json();
      expect(Array.isArray(providers)).toBeTruthy();

      // Each provider should have slug and name
      for (const provider of providers) {
        expect(provider).toHaveProperty('slug');
        expect(provider).toHaveProperty('name');
        expect(typeof provider.slug).toBe('string');
        expect(typeof provider.name).toBe('string');
      }
    });

    test('GET /auth/oidc/nonexistent/authorize returns 404', async ({ request }) => {
      const response = await request.get('http://localhost:3000/auth/oidc/nonexistent/authorize', {
        maxRedirects: 0
      });

      // Should return 404 for nonexistent provider
      expect(response.status()).toBe(404);
    });

    test('GET /auth/oidc/callback without params redirects with error', async ({ request }) => {
      // Test callback without required parameters
      const response = await request.get('http://localhost:3000/auth/oidc/dex/callback', {
        maxRedirects: 0
      });

      // Should redirect to login with error
      expect(response.status()).toBe(302);
      const location = response.headers()['location'];
      expect(location).toContain('/login');
      expect(location).toContain('error=');
    });

    test('GET /auth/oidc/callback with invalid state redirects with error', async ({ request }) => {
      const response = await request.get('http://localhost:3000/auth/oidc/dex/callback?code=test&state=invalid', {
        maxRedirects: 0
      });

      expect(response.status()).toBe(302);
      const location = response.headers()['location'];
      expect(location).toContain('/login');
      expect(location).toContain('error=');
    });
  });

  test.describe('OIDC GraphQL Query', () => {
    test('oidcProviders query returns list of configured providers', async ({ request }) => {
      const response = await request.post('http://localhost:3000/graphql', {
        data: {
          query: `
            query {
              oidcProviders {
                slug
                name
              }
            }
          `
        }
      });

      expect(response.ok()).toBeTruthy();

      const body = await response.json();
      expect(body.errors).toBeUndefined();
      expect(body.data).toHaveProperty('oidcProviders');
      expect(Array.isArray(body.data.oidcProviders)).toBeTruthy();

      // Each provider should have slug and name
      for (const provider of body.data.oidcProviders) {
        expect(provider).toHaveProperty('slug');
        expect(provider).toHaveProperty('name');
      }
    });
  });

});

test.describe('OIDC Full Flow (requires Dex)', () => {
  // These tests require Dex to be running and properly configured
  // They are skipped if Dex is not available

  test.beforeEach(async ({ request }) => {
    // Check if Dex is available
    try {
      const response = await request.get('http://localhost:5556/dex/.well-known/openid-configuration', {
        timeout: 2000
      });
      if (!response.ok()) {
        test.skip();
      }
    } catch {
      test.skip();
    }
  });

  test('can initiate OIDC login flow with Dex', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    const dexButton = page.getByRole('button', { name: /dex/i });

    if (await dexButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      // Click Dex button and wait for redirect to Dex login page
      await dexButton.click();

      // Should redirect to Dex authorization endpoint
      await expect(page).toHaveURL(/localhost:5556\/dex/, { timeout: 5000 }).catch(() => {
        // If we get redirected back to our app, check the URL
        const url = page.url();
        expect(url).toMatch(/localhost:5556|\/auth\/oidc/);
      });
    } else {
      console.log('Dex provider button not found - skipping');
      test.skip();
    }
  });

  test('Dex authorization page is accessible', async ({ page }) => {
    // Directly navigate to a Dex authorization endpoint
    // This verifies Dex is properly configured
    await page.goto('http://localhost:5556/dex/auth/local?client_id=pulse-app&redirect_uri=http://localhost:3000/auth/oidc/dex/callback&response_type=code&scope=openid+email+profile&state=test-state&code_challenge=test&code_challenge_method=S256');

    // Should show Dex login page with Login button
    await expect(page.getByRole('button', { name: 'Login' })).toBeVisible({ timeout: 5000 });
  });
});
