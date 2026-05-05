import { test, expect } from '@playwright/test';
import { loginAsAdmin, createEvent } from '../fixtures/index';

test.describe('Events + RSVP', () => {
  test('dashboard shows upcoming events with own RSVP state', async ({ page }) => {
    await loginAsAdmin(page);

    const eventName = `Dashboard-Event-${Date.now()}`;
    await createEvent(page, { name: eventName, type: 'rehearsal' });

    await page.goto('/tableau-de-bord');
    await expect(page.locator(`a:has-text("${eventName}")`)).toBeVisible();

    // RSVP state column must be present
    await expect(page.locator('.rsvp-state, td:has-text("Sans réponse"), td:has-text("Présent")').first()).toBeVisible();
  });

  test('concert RSVP form includes instrument dropdown', async ({ page }) => {
    await loginAsAdmin(page);
    const { id } = await createEvent(page, { name: `Concert-${Date.now()}`, type: 'concert' });

    await page.goto(`/evenements/${id}`);
    // Select "Présent" to reveal the instrument dropdown (radio input is visually hidden; click label)
    await page.click('label[for="rsvp-yes"]');
    await expect(page.locator('#rsvp-instrument')).toBeVisible();
  });

  test('rehearsal RSVP form has no instrument dropdown', async ({ page }) => {
    await loginAsAdmin(page);
    const { id } = await createEvent(page, { name: `Repet-${Date.now()}`, type: 'rehearsal' });

    await page.goto(`/evenements/${id}`);
    await expect(page.locator('#rsvp-instrument')).not.toBeAttached();
  });

  test('custom field lifecycle: add, edit, delete without responses', async ({ page }) => {
    await loginAsAdmin(page);
    const { id } = await createEvent(page, { name: `Custom-Fields-${Date.now()}`, type: 'other' });

    // Navigate to event edit page
    await page.goto(`/admin/evenements/${id}/modifier`);

    // Expand the "Add field" details panel
    await page.click('details summary');

    // Add a field
    const fieldLabel = `Champ-${Date.now()}`;
    await page.fill('#field-label', fieldLabel);
    await page.click('form[action*="/champs"] [type=submit]');

    // Field appears in the table
    await expect(page.locator(`td:has-text("${fieldLabel}")`)).toBeVisible();

    // Extract field edit URL
    const editLink = page.locator(`tr:has-text("${fieldLabel}") a:has-text("Modifier")`);
    const editHref = await editLink.getAttribute('href');
    expect(editHref).toBeTruthy();

    // Edit the field label
    await page.goto(editHref!);
    const updatedLabel = `${fieldLabel}-updated`;
    await page.fill('#field-label', updatedLabel);
    await page.click('[type=submit]');

    // Updated label appears on edit page
    await expect(page.locator(`td:has-text("${updatedLabel}")`)).toBeVisible();

    // Delete the field
    const deleteBtn = page.locator(`tr:has-text("${updatedLabel}") button:has-text("Supprimer")`);
    await deleteBtn.click();
    await expect(page.locator('#confirm-modal')).toBeVisible();
    await page.click('#confirm-modal button[type=submit]');

    // Field no longer present
    await expect(page.locator(`td:has-text("${updatedLabel}")`)).not.toBeVisible();
    // No danger flash (delete succeeded)
    await expect(page.locator('.alert--danger')).not.toBeVisible();
  });

  test('field edit blocked when responses exist', async ({ page }) => {
    await loginAsAdmin(page);
    const eventName = `Blocked-Edit-${Date.now()}`;
    const { id } = await createEvent(page, { name: eventName, type: 'other' });

    // Add a custom field
    await page.goto(`/admin/evenements/${id}/modifier`);
    await page.click('details summary');
    const fieldLabel = `BlockedField-${Date.now()}`;
    await page.fill('#field-label', fieldLabel);
    await page.click('form[action*="/champs"] [type=submit]');
    await expect(page.locator(`td:has-text("${fieldLabel}")`)).toBeVisible();

    // Submit an RSVP with a field response as the admin (who is also a member)
    await page.goto(`/evenements/${id}`);
    await page.click('label[for="rsvp-yes"]');
    // Fill the custom field if visible
    const fieldInputs = page.locator('[name^="field_"]');
    const fieldCount = await fieldInputs.count();
    if (fieldCount > 0) {
      await fieldInputs.first().fill('réponse test');
    }
    await page.click('form[action*="/rsvp"] [type=submit]');

    // Now try to edit the field as admin
    await page.goto(`/admin/evenements/${id}/modifier`);
    const editLink = page.locator(`tr:has-text("${fieldLabel}") a:has-text("Modifier")`);
    const editHref = await editLink.getAttribute('href');
    await page.goto(editHref!);
    await page.fill('#field-label', 'new-label');
    await page.click('[type=submit]');

    // A danger flash message must appear indicating the edit is blocked
    await expect(page.locator('.alert--danger')).toBeVisible();
    const alertText = await page.locator('.alert--danger').textContent();
    expect(alertText).toContain('réponses');
  });
});
