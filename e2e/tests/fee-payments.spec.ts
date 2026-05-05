import { test, expect } from '@playwright/test';
import { loginAsAdmin, createMusician, createSeason } from '../fixtures/index';

test.describe('Fee Payments', () => {
  test('fee payment section renders on musician detail', async ({ page }) => {
    await loginAsAdmin(page);
    const { id } = await createMusician(page, {
      firstName: 'Fee',
      lastName: `Section${Date.now()}`,
      email: `fee-section-${Date.now()}@example.com`,
    });

    await page.goto(`/admin/musiciens/${id}`);

    // The Cotisations section must be present
    await expect(page.locator('section:has(h2:has-text("Cotisations"))')).toBeVisible();
    // The "Enregistrer" button to add a payment must be visible
    await expect(page.locator('button:has-text("+ Enregistrer")')).toBeVisible();
  });

  test('earlier payment updates first inscription date', async ({ page }) => {
    await loginAsAdmin(page);

    // Two seasons needed: fee_payments has UNIQUE(account_id, season_id), so each payment
    // must target a distinct season to avoid a duplicate key error.
    const ts = Date.now();
    const season2024Label = `Fee-Season-2024-${ts}`;
    const season2023Label = `Fee-Season-2023-${ts}`;
    await createSeason(page, season2024Label, { startDate: '2024-09-01', endDate: '2025-06-30' });
    await createSeason(page, season2023Label, { startDate: '2023-09-01', endDate: '2024-06-30' });

    const { id } = await createMusician(page, {
      firstName: 'Fee',
      lastName: `Date${ts}`,
      email: `fee-date-${ts}@example.com`,
    });

    // Add first payment with a later date (2024 season)
    await page.goto(`/admin/musiciens/${id}`);
    await page.click('button:has-text("+ Enregistrer")');
    await expect(page.locator('#fee-payment-modal')).toBeVisible();
    await page.selectOption('#cotisation-season-id', { label: season2024Label });
    await page.selectOption('#cotisation-payment-type', { index: 1 });
    await page.fill('#cotisation-amount', '50');
    await page.fill('#cotisation-payment-date', '2024-10-15');
    await page.click('#fee-payment-modal [type=submit]');
    await page.waitForURL(`**/admin/musiciens/${id}`);

    const firstDateText1 = await page
      .locator('dt:has-text("Date de première inscription") + dd')
      .textContent();
    expect(firstDateText1?.trim()).toBeTruthy();

    // Add second payment with an earlier date (2023 season)
    await page.click('button:has-text("+ Enregistrer")');
    await expect(page.locator('#fee-payment-modal')).toBeVisible();
    await page.selectOption('#cotisation-season-id', { label: season2023Label });
    await page.selectOption('#cotisation-payment-type', { index: 1 });
    await page.fill('#cotisation-amount', '30');
    await page.fill('#cotisation-payment-date', '2023-09-01');
    await page.click('#fee-payment-modal [type=submit]');
    await page.waitForURL(`**/admin/musiciens/${id}`);

    // First inscription date must now reflect the earlier payment
    await page.goto(`/admin/musiciens/${id}`);
    const firstDateText2 = await page
      .locator('dt:has-text("Date de première inscription") + dd')
      .textContent();
    expect(firstDateText2).toContain('2023');
  });
});
