import { test, expect } from '@playwright/test';
import { loginAsAdmin, logout } from '../fixtures/index';

test.describe('Navigation', () => {
  test('unauthenticated nav shows only login link', async ({ page }) => {
    await logout(page);
    await page.goto('/');
    await expect(page.locator('a[href="/connexion"]').first()).toBeVisible();
    await expect(page.locator('a[href="/tableau-de-bord"]')).not.toBeVisible();
    await expect(page.locator('a[href="/evenements"]')).not.toBeVisible();
    await expect(page.locator('a[href="/deconnexion"]')).not.toBeVisible();
  });

  test('authenticated nav shows full menu', async ({ page }) => {
    await loginAsAdmin(page);
    await expect(page.locator('a[href="/tableau-de-bord"]').first()).toBeVisible();
    await expect(page.locator('a[href="/evenements"]').first()).toBeVisible();
    await expect(page.locator('a[href="/profil"]').first()).toBeVisible();
    await expect(page.locator('a[href="/deconnexion"]').first()).toBeVisible();
  });

  test('privacy notice page is not blank', async ({ page }) => {
    await page.goto('/politique-de-confidentialite');
    const body = await page.locator('body').textContent();
    expect(body?.length).toBeGreaterThan(100);
    // Page should render more than a handful of words
    const textLength = (body ?? '').replace(/\s+/g, ' ').trim().length;
    expect(textLength).toBeGreaterThan(200);
  });
});
