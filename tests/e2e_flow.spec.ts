import { test, expect } from '@playwright/test'

const base = process.env.E2E_BASE || 'http://127.0.0.1:28353'

test.describe('GotoSky critical paths', () => {
  test('health via same origin', async ({ request }) => {
    const r = await request.get(base + '/health')
    expect(r.ok()).toBeTruthy()
    const j = await r.json()
    expect(j.status).toBe('ok')
  })

  test('calendar gold/silver/bronze and drag target', async ({ page }) => {
    await page.goto(base + '/')
    await expect(page.getByText('24 小时夜间日历')).toBeVisible()
    await page.getByTestId('refresh-weather').click()
    await page.waitForTimeout(1500)
    await expect(page.getByTestId('night-strip')).toBeVisible()
    const src = page.getByTestId('target-list').locator('div').first()
    const dst = page.getByTestId('drop-plan')
    await src.dragTo(dst)
    await expect(page.getByText(/已将天体拖入/)).toBeVisible({ timeout: 8000 })
  })

  test('heatmap layers', async ({ page }) => {
    await page.goto(base + '/heatmap')
    await expect(page.getByText('天气热力矩阵')).toBeVisible()
    await page.getByRole('button', { name: '视宁度' }).click()
    await expect(page.getByTestId('heatmap')).toBeVisible()
  })

  test('sky canvas', async ({ page }) => {
    await page.goto(base + '/sky')
    await expect(page.getByTestId('sky-canvas')).toBeVisible()
  })

  test('rig wall start session', async ({ page }) => {
    await page.goto(base + '/wall')
    await expect(page.getByTestId('rig-wall')).toBeVisible()
    const btn = page.getByRole('button', { name: '启动追星' }).first()
    if (await btn.isVisible()) {
      await btn.click()
      await expect(page.getByText(/会话已启动|IDLE|CONNECTING|SLEWING|EXPOSING/)).toBeVisible({ timeout: 8000 })
    }
  })
})
