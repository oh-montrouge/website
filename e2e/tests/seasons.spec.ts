import { test, expect } from '@playwright/test';
import { loginAsAdmin, createSeason } from '../fixtures/index';

test.describe('Season Management', () => {
  test('current season is visually distinguished', async ({ page }) => {
    await loginAsAdmin(page);

    // Create two seasons and designate one as current
    const label1 = `S-A-${Date.now()}`;
    const label2 = `S-B-${Date.now()}`;
    await createSeason(page, label1);
    await createSeason(page, label2);

    // Designate label1 as current
    await page.goto('/admin/saisons');
    const row = page.locator(`tr:has-text("${label1}")`);
    await row.locator('button:has-text("Désigner comme courante")').click();

    // After redirect, label1's row must have the "Saison courante" badge
    await expect(
      page.locator(`tr:has-text("${label1}") .badge:has-text("Saison courante")`),
    ).toBeVisible();

    // label1's row must NOT have a "Désigner" button (it is already current)
    await expect(
      page.locator(`tr:has-text("${label1}") button:has-text("Désigner comme courante")`),
    ).not.toBeVisible();
  });

  test('designating a season updates immediately without manual reload', async ({ page }) => {
    await loginAsAdmin(page);

    const label = `S-Imm-${Date.now()}`;
    await createSeason(page, label);

    await page.goto('/admin/saisons');
    const row = page.locator(`tr:has-text("${label}")`);
    const designateBtn = row.locator('button:has-text("Désigner comme courante")');

    // Button must be visible before click
    await expect(designateBtn).toBeVisible();
    await designateBtn.click();

    // After form submission (full page reload) the badge appears immediately
    await expect(
      page.locator(`tr:has-text("${label}") .badge:has-text("Saison courante")`),
    ).toBeVisible();
  });
});
