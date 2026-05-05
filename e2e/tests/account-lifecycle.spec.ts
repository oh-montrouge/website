import { test, expect } from '@playwright/test';
import { loginAsAdmin, createMusician, logout } from '../fixtures/index';

test.describe('Account Lifecycle', () => {
  test('invite form pre-populates account email', async ({ page }) => {
    await loginAsAdmin(page);
    const email = `invite-test-${Date.now()}@example.com`;
    const { id } = await createMusician(page, {
      firstName: 'Invite',
      lastName: 'Test',
      email,
    });

    // Musician page shows an existing invite link (created on account creation)
    // or we can generate a new one
    await page.goto(`/admin/musiciens/${id}`);
    await page.click('button:has-text("Afficher le lien")');
    const inviteInput = page.locator('input[readonly][aria-label="URL d\'invitation"]');
    await expect(inviteInput).toBeVisible();
    const inviteURL = await inviteInput.inputValue();
    expect(inviteURL).toMatch(/\/invitation\//);

    // Follow the invite link; the form must show the musician's email
    await logout(page);
    await page.goto(inviteURL);
    await expect(page.locator(`text=${email}`)).toBeVisible();
  });

  test('password reset form is reachable', async ({ page }) => {
    await loginAsAdmin(page);
    const email = `reset-reach-${Date.now()}@example.com`;
    const { id } = await createMusician(page, {
      firstName: 'Reset',
      lastName: 'Reach',
      email,
    });

    // Complete the invite first so the account becomes active (reset only works for active accounts)
    await page.goto(`/admin/musiciens/${id}`);
    await page.click('button:has-text("Afficher le lien")');
    const inviteInput = page.locator('input[readonly][aria-label="URL d\'invitation"]');
    const inviteURL = await inviteInput.inputValue();

    await logout(page);
    await page.goto(inviteURL);
    await page.fill('#password', 'InitialPass123!SecretX@');
    await page.fill('#password_confirm', 'InitialPass123!SecretX@');
    await page.check('[name=privacy_consent]');
    await page.click('[type=submit]');
    await expect(page).not.toHaveURL(/\/invitation\//);

    // Back as admin: generate a reset link
    await loginAsAdmin(page);
    await page.goto(`/admin/musiciens/${id}`);
    await page.click('form[action*="/reinitialisation"] button[type=submit]');
    // Page reloads with a reset token URL; reveal it (hidden by Alpine.js by default)
    await page.click('button:has-text("Afficher le lien")');
    const resetInput = page.locator('input[readonly][aria-label="URL de réinitialisation"]');
    await expect(resetInput).toBeVisible();
    const resetURL = await resetInput.inputValue();
    expect(resetURL).toMatch(/\/reinitialiser-mot-de-passe\//);

    await logout(page);
    await page.goto(resetURL);
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('#password_confirm')).toBeVisible();
  });

  test('password reset completes successfully', async ({ page }) => {
    await loginAsAdmin(page);
    const email = `reset-complete-${Date.now()}@example.com`;
    const { id } = await createMusician(page, {
      firstName: 'Reset',
      lastName: 'Complete',
      email,
    });

    // Activate account via invite
    await page.goto(`/admin/musiciens/${id}`);
    await page.click('button:has-text("Afficher le lien")');
    const inviteInput = page.locator('input[readonly][aria-label="URL d\'invitation"]');
    const inviteURL = await inviteInput.inputValue();

    await logout(page);
    await page.goto(inviteURL);
    await page.fill('#password', 'InitialPass123!SecretX@');
    await page.fill('#password_confirm', 'InitialPass123!SecretX@');
    await page.check('[name=privacy_consent]');
    await page.click('[type=submit]');

    // Generate reset link as admin
    await loginAsAdmin(page);
    await page.goto(`/admin/musiciens/${id}`);
    await page.click('form[action*="/reinitialisation"] button[type=submit]');
    await page.click('button:has-text("Afficher le lien")');
    const resetInput = page.locator('input[readonly][aria-label="URL de réinitialisation"]');
    await expect(resetInput).toBeVisible();
    const resetURL = await resetInput.inputValue();

    await logout(page);
    await page.goto(resetURL);
    await page.fill('#password', 'ResetPassword456!Valid@');
    await page.fill('#password_confirm', 'ResetPassword456!Valid@');
    await page.click('[type=submit]');

    // On success, redirected to /connexion
    await expect(page).toHaveURL(/\/connexion/);
  });
});
