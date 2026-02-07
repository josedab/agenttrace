import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('redirects unauthenticated users to sign-in', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/sign-in/);
  });

  test('sign-in page renders correctly', async ({ page }) => {
    await page.goto('/sign-in');
    await expect(page.locator('h1, h2, [role="heading"]').first()).toBeVisible();
  });
});

test.describe('Health Check', () => {
  test('sign-in page loads without errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (err) => errors.push(err.message));

    await page.goto('/sign-in');
    await page.waitForLoadState('networkidle');

    expect(errors).toHaveLength(0);
  });
});
