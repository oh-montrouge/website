import { Page, expect } from '@playwright/test';

const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@e2e.local';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'testpassword';

export async function loginAsAdmin(page: Page): Promise<void> {
  await page.goto('/connexion');
  await page.fill('#email', ADMIN_EMAIL);
  await page.fill('#password', ADMIN_PASSWORD);
  await page.click('[type=submit]');
  await expect(page).not.toHaveURL('/connexion');
}

export async function logout(page: Page): Promise<void> {
  await page.goto('/deconnexion');
}

export interface MusicianOptions {
  firstName: string;
  lastName: string;
  email: string;
}

export async function createMusician(
  page: Page,
  opts: MusicianOptions,
): Promise<{ id: string }> {
  await page.goto('/admin/musiciens/nouveau');
  await page.fill('#first_name', opts.firstName);
  await page.fill('#last_name', opts.lastName);
  await page.fill('#email', opts.email);
  // Pick first available instrument
  await page.selectOption('#main_instrument_id', { index: 1 });
  await page.click('[type=submit]');
  // After creation, we are on the musician detail page
  const url = page.url();
  const id = url.match(/\/admin\/musiciens\/(\d+)/)?.[1] ?? '';
  return { id };
}

export interface EventOptions {
  name: string;
  type: 'concert' | 'rehearsal' | 'other';
  date?: string; // YYYY-MM-DD, defaults to a future date
  time?: string; // HH:MM, defaults to 20:00
  description?: string;
}

export async function createEvent(
  page: Page,
  opts: EventOptions,
): Promise<{ id: string }> {
  await page.goto('/admin/evenements/nouveau');
  await page.fill('#event-name', opts.name);
  const date = opts.date ?? futureDate();
  await page.fill('#event-date', date);
  await page.fill('#event-time', opts.time ?? '20:00');
  await page.selectOption('#event-type', opts.type);
  if (opts.description) {
    await page.fill('#event-description', opts.description);
  }
  await page.click('[type=submit]');
  // After creation, navigate to the member events list and find the event by name
  await page.goto('/evenements');
  const eventLink = page.locator(`a:has-text("${opts.name}")`).first();
  await expect(eventLink).toBeVisible({ timeout: 5000 });
  const href = await eventLink.getAttribute('href');
  const id = href?.match(/\/evenements\/(\d+)/)?.[1] ?? '';
  return { id };
}

export interface SeasonOptions {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export async function createSeason(
  page: Page,
  label: string,
  options: SeasonOptions = {},
): Promise<void> {
  await page.goto('/admin/saisons');
  await page.click('button:has-text("+ Nouvelle saison")');
  // Wait for the dialog to open (Alpine.js calls showModal() which sets the [open] attribute)
  await expect(page.locator('dialog[open]').first()).toBeVisible({ timeout: 3000 });
  await page.fill('#label', label);
  const year = new Date().getFullYear();
  await page.fill('#start_date', options.startDate ?? `${year}-09-01`);
  await page.fill('#end_date', options.endDate ?? `${year + 1}-06-30`);
  await page.click('dialog [type=submit]');
  // Wait for the form submission to complete (redirects back to the seasons list)
  await expect(page.locator(`td:has-text("${label}")`)).toBeVisible({ timeout: 5000 });
}

function futureDate(): string {
  const d = new Date();
  d.setDate(d.getDate() + 30);
  return d.toISOString().split('T')[0];
}
