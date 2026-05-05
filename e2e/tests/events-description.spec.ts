import { test, expect } from '@playwright/test';
import { loginAsAdmin, createEvent } from '../fixtures/index';

test.describe('Events — dashboard and description', () => {
  test('dashboard renders card layout, not table', async ({ page }) => {
    await loginAsAdmin(page);
    await createEvent(page, { name: `Card-Test-${Date.now()}`, type: 'rehearsal' });

    await page.goto('/tableau-de-bord');
    await expect(page.locator('.card').first()).toBeVisible();
    await expect(page.locator('[aria-label="Liste des événements"]')).not.toBeAttached();
  });

  test('markdown description renders as HTML on dashboard and detail page', async ({ page }) => {
    await loginAsAdmin(page);
    const eventName = `Markdown-${Date.now()}`;
    const { id } = await createEvent(page, {
      name: eventName,
      type: 'rehearsal',
      description: '**Gala** de printemps',
    });

    await page.goto('/tableau-de-bord');
    await expect(page.locator('.card strong:has-text("Gala")')).toBeVisible();

    await page.goto(`/evenements/${id}`);
    await expect(page.locator('.prose strong:has-text("Gala")')).toBeVisible();
  });

  test('admin can set and update description', async ({ page }) => {
    await loginAsAdmin(page);
    const eventName = `Edit-Desc-${Date.now()}`;
    const { id } = await createEvent(page, {
      name: eventName,
      type: 'rehearsal',
      description: 'Description initiale',
    });

    await page.goto(`/evenements/${id}`);
    await expect(page.locator('.prose')).toContainText('Description initiale');

    await page.goto(`/admin/evenements/${id}/modifier`);
    await expect(page.locator('#event-description')).toHaveValue('Description initiale');
    await page.fill('#event-description', 'Description **mise à jour**');
    await page.click('[type=submit]');

    await page.goto(`/evenements/${id}`);
    await expect(page.locator('.prose strong:has-text("mise à jour")')).toBeVisible();
  });

  test('empty description does not render description block', async ({ page }) => {
    await loginAsAdmin(page);
    const { id } = await createEvent(page, { name: `No-Desc-${Date.now()}`, type: 'concert' });

    await page.goto(`/evenements/${id}`);
    await expect(page.locator('.prose')).not.toBeAttached();
  });

  test('events list renders table layout after dashboard split', async ({ page }) => {
    await loginAsAdmin(page);
    const eventName = `List-Test-${Date.now()}`;
    await createEvent(page, { name: eventName, type: 'concert' });

    await page.goto('/evenements');
    await expect(page.locator('[aria-label="Liste des événements"]')).toBeVisible();
    await expect(page.locator(`a:has-text("${eventName}")`)).toBeVisible();
  });
});
