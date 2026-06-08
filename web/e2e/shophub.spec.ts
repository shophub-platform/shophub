import { test, expect, Page } from '@playwright/test';

// Happy-path E2E za ShopHub UI (FZ 4, 9.4).
// API se mokuje preko Playwright route-ova, pa test ne zavisi od backend-a.

const tokens = { accessToken: 'a.b.c', refreshToken: 'r.e.f', tokenType: 'Bearer' };

const sampleShop = {
  id: 'shop-1',
  name: 'moja-radnja',
  availability: 'high',
  walletAddress: '0xabc',
  databaseType: 'postgres',
  image: 'ghcr.io/shophub-platform/shop:latest',
  status: 'running',
  url: 'http://shop-1.shophub.local',
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

async function mockApi(page: Page, shops: unknown[]): Promise<void> {
  await page.route('**/api/v1/auth/register', (route) =>
    route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify(tokens) }),
  );
  await page.route('**/api/v1/auth/login', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(tokens) }),
  );
  await page.route('**/api/v1/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ userID: 'u1', email: 'vesna@b.com' }),
    }),
  );
  await page.route('**/api/v1/shops', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(shops) });
    }
    // POST (kreiranje)
    return route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify(sampleShop) });
  });
}

test('registracija vodi na dashboard sa listom prodavnica', async ({ page }) => {
  await mockApi(page, [sampleShop]);

  await page.goto('/register');
  await page.getByLabel('Display name').fill('Vesna');
  await page.getByLabel('Email').fill('vesna@b.com');
  await page.getByLabel('Password').fill('supersecret');
  await page.getByRole('button', { name: /Create account/ }).click();

  await expect(page.getByRole('heading', { name: 'My shops' })).toBeVisible();
  await expect(page.getByText('moja-radnja')).toBeVisible();
  await expect(page.getByText('Ready')).toBeVisible();
});

test('prazan dashboard prikazuje poziv na kreiranje', async ({ page }) => {
  await mockApi(page, []);

  await page.goto('/login');
  await page.getByLabel('Email').fill('vesna@b.com');
  await page.getByLabel('Password').fill('supersecret');
  await page.getByRole('button', { name: /Sign in/ }).click();

  await expect(page.getByText("You don't have any shops yet.")).toBeVisible();
  await expect(page.getByRole('link', { name: 'Create your first' })).toBeVisible();
});

test('wizard kreira prodavnicu kroz 3 koraka', async ({ page }) => {
  await mockApi(page, []);

  // Prijava
  await page.goto('/login');
  await page.getByLabel('Email').fill('vesna@b.com');
  await page.getByLabel('Password').fill('supersecret');
  await page.getByRole('button', { name: /Sign in/ }).click();
  await expect(page.getByRole('heading', { name: 'My shops' })).toBeVisible();

  // Wizard
  await page.goto('/shops/new');
  // Korak 1 -> 2 (prvo "Dalje")
  await page.getByLabel('Shop name').fill('moja-radnja');
  await page.getByRole('button', { name: 'Next' }).first().click();
  // Korak 2 -> 3 (drugo "Dalje"); vertikalni stepper drži oba dugmeta u DOM-u
  await page.getByLabel('Wallet address').fill('0xabc');
  await page.getByRole('button', { name: 'Next' }).nth(1).click();
  await page.getByRole('button', { name: /Create shop/ }).click();

  // Posle kreiranja vodi na detalj
  await expect(page).toHaveURL(/\/shops\/shop-1/);
});
