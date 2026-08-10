import { test, expect } from '@playwright/test'

test('user can create a PostgreSQL instance', async ({ page }) => {
  await page.goto('/instances/new')

  await page.getByLabel('Instance name').fill('my-db-2')

  await page.getByLabel('PostgreSQL instances').fill('2')

  await page.getByLabel('PostgreSQL version').fill('15')

  await page.getByLabel('Storage').fill('7Gi')

  await page
    .getByRole('button', {
      name: 'Create instance',
    })
    .click()

  await expect(page).toHaveURL(/\/instances\/inst-/)
})
