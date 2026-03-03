import { test, expect } from '@playwright/test';

test.describe('Travel UI Tests', () => {
    test.beforeEach(async ({ page }) => {
        // Navigate to the travel page
        await page.goto('http://127.0.0.1:8080/travel');
    });

    test('Desktop: Tabs and Sidebar', async ({ page }) => {
        // Set viewport for desktop
        await page.setViewportSize({ width: 1280, height: 720 });

        // 1. Verify tabs are present and have the correct class
        const tabs = page.locator('.btn-tab');
        await expect(tabs).toHaveCount(8); // 2018 to 2025

        // 2. Verify selecting a different year updates the table
        const year2023Tab = page.locator('button:has-text("2023")');
        await year2023Tab.click();
        await expect(page.locator('h2')).toContainText('2023');

        // 3. Verify desktop sidebar is visible and autoloads
        const sidebar = page.locator('.md\\:block.md\\:w-1\/4');
        await expect(sidebar).toBeVisible();
        await expect(sidebar).toContainText('Top Destinations');
    });

    test('Mobile: Lazy-loaded Dropdown', async ({ page }) => {
        // Set viewport for mobile
        await page.setViewportSize({ width: 375, height: 667 });

        // 1. Verify desktop sidebar is hidden
        const sidebar = page.locator('.md\\:block.md\\:w-1\/4');
        await expect(sidebar).toBeHidden();

        // 2. Verify mobile dropdown is present
        const dropdown = page.locator('details');
        await expect(dropdown).toBeVisible();
        await expect(dropdown.locator('summary')).toContainText('Top Destinations');

        // 3. Verify content is lazy-loaded on click
        await dropdown.locator('summary').click();
        // Wait for HTMX to load content
        const mobileContent = page.locator('#top-destinations-mobile');
        await expect(mobileContent).toContainText('Top Destinations', { timeout: 10000 });
    });
});
