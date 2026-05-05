import { test, expect } from '@playwright/test';
import { loginAsAdmin, createMusician, logout } from '../fixtures/index';

test.describe('Musician Management', () => {
  test('invite URL leads to complete-invite form', async ({ page }) => {
    await loginAsAdmin(page);
    const { id } = await createMusician(page, {
      firstName: 'Invite',
      lastName: 'Flow',
      email: `invite-flow-${Date.now()}@example.com`,
    });

    await page.goto(`/admin/musiciens/${id}`);
    await page.click('button:has-text("Afficher le lien")');
    const inviteURL = await page
      .locator('input[readonly][aria-label="URL d\'invitation"]')
      .inputValue();

    await logout(page);
    await page.goto(inviteURL);

    // Form must be present with password and privacy consent fields
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('[name=privacy_consent]')).toBeVisible();
  });

  test('musician detail page shows all key sections', async ({ page }) => {
    await loginAsAdmin(page);
    const firstName = 'Detail';
    const lastName = `Musicien${Date.now()}`;
    const { id } = await createMusician(page, {
      firstName,
      lastName,
      email: `detail-${Date.now()}@example.com`,
    });

    await page.goto(`/admin/musiciens/${id}`);

    // Name appears in header
    await expect(page.locator(`h1:has-text("${lastName}")`)).toBeVisible();

    // Profile section
    await expect(page.locator('section:has(h2:has-text("Profil"))')).toBeVisible();

    // Status badge (pending because no invite completed)
    await expect(page.locator('.badge:has-text("En attente"), .badge:has-text("Actif")')).toBeVisible();

    // Fee payments section
    await expect(page.locator('section:has(h2:has-text("Cotisations"))')).toBeVisible();
  });

  test('anonymization removes identifying information', async ({ page }) => {
    await loginAsAdmin(page);
    const firstName = 'Efface';
    const lastName = `ToErase${Date.now()}`;
    const email = `anon-${Date.now()}@example.com`;
    const { id } = await createMusician(page, { firstName, lastName, email });

    await page.goto(`/admin/musiciens/${id}`);

    // Open the Alpine confirm modal
    await page.click('button:has-text("Anonymiser ce compte")');

    // The custom confirm modal requires typing "ANONYMISER"
    await expect(page.locator('#confirm-modal')).toBeVisible();
    await page.fill('#confirm-modal-input', 'ANONYMISER');
    await page.click('#confirm-modal button[type=submit]');

    // After anonymization the page should no longer show the original name/email
    await page.waitForURL(/\/admin\/musiciens\/\d+/);
    const bodyText = await page.locator('body').textContent();
    expect(bodyText).not.toContain(firstName);
    expect(bodyText).not.toContain(email);

    // Status badge must show anonymized state
    await expect(page.locator('.badge:has-text("Anonymisé")')).toBeVisible();
  });
});
