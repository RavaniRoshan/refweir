#!/usr/bin/env node
import { existsSync, readFileSync, mkdirSync, writeFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'
import { createRequire } from 'node:module'

const __dirname = dirname(fileURLToPath(import.meta.url))
const ROOT_DIR = resolve(__dirname, '../..')
const SITE_PRODUCT_DIR = resolve(ROOT_DIR, 'site/public/product')
const SITE_PUBLIC_DIR = resolve(ROOT_DIR, 'site/public')

// Resolve playwright from web/node_modules
let chromium
try {
  const req = createRequire(resolve(ROOT_DIR, 'web/package.json'))
  chromium = req('playwright').chromium
} catch {
  const pw = await import('playwright')
  chromium = pw.chromium
}

const workDir = process.env.REFWEIR_CAPTURE_DIR || '/tmp/refweir-capture'
const port = process.env.PORT || '18080'

// 1. Ensure instance is seeded and running
async function ensureServerRunning() {
  try {
    const res = await fetch(`http://127.0.0.1:${port}/healthz`)
    if (res.ok) return
  } catch {}

  console.log(`Starting seeded production instance at http://127.0.0.1:${port}...`)
  execSync(`bash "${resolve(__dirname, 'seed.sh')}" "${workDir}" "${port}"`, {
    stdio: 'inherit',
    cwd: ROOT_DIR,
  })

  // Wait for health check
  for (let i = 0; i < 40; i++) {
    try {
      const res = await fetch(`http://127.0.0.1:${port}/healthz`)
      if (res.ok) {
        console.log(`Instance ready on http://127.0.0.1:${port}`)
        return
      }
    } catch {}
    await new Promise((r) => setTimeout(r, 250))
  }
  throw new Error(`Server failed to start on port ${port}`)
}

await ensureServerRunning()

const tokenFile = resolve(workDir, 'key')
if (!existsSync(tokenFile)) {
  throw new Error(`Token file not found at ${tokenFile}`)
}
const token = readFileSync(tokenFile, 'utf8').trim()

mkdirSync(SITE_PRODUCT_DIR, { recursive: true })

// 2. Launch Playwright browser with production specifications (1440x900 @2x, dark)
const browser = await chromium.launch({
  executablePath:
    process.env.CHROMIUM_PATH ||
    (existsSync('/usr/bin/chromium-browser') ? '/usr/bin/chromium-browser' : undefined),
  args: ['--no-sandbox', '--disable-setuid-sandbox'],
})

const context = await browser.newContext({
  viewport: { width: 1440, height: 900 },
  deviceScaleFactor: 2,
  colorScheme: 'dark',
})

// Inject credentials and preferences into localStorage
await context.addInitScript(({ token }) => {
  window.localStorage.setItem('refweir_token', token)
  window.localStorage.setItem('refweir_onboarding_tour_completed', 'true')
  window.localStorage.setItem('refweir_theme', 'dark')
  window.localStorage.setItem('refweir-theme', 'dark')
  window.localStorage.setItem('refweir_demo_banner_collapsed', 'true')
  window.localStorage.setItem('refweir.character_shortcuts_enabled', 'false')
}, { token })

const page = await context.newPage()

// Strict assertion helper: fail if TanStack Router devtools or any dialog exists
async function assertCleanState(stepName) {
  const routerTextLocator = page.locator('text="TanStack Router"')
  if ((await routerTextLocator.count()) > 0) {
    throw new Error(`[${stepName}] Assertion Failed: Found TanStack Router devtools element on page.`)
  }

  const dialogLocator = page.locator('[role="dialog"]')
  if ((await dialogLocator.count()) > 0) {
    const dialogHtml = await dialogLocator.first().innerHTML().catch(() => '')
    throw new Error(`[${stepName}] Assertion Failed: Found [role="dialog"] modal on page: ${dialogHtml.slice(0, 100)}`)
  }
}

async function safeGoto(url) {
  await page.goto(url)
  await page.waitForLoadState('networkidle', { timeout: 4000 }).catch(() => {})
}

console.log('Capturing 5 production screenshots (1440x900 @2x, dark)...')

// View 1: Repository Overview
console.log('1/5: Capturing repo overview...')
await safeGoto(`http://127.0.0.1:${port}/?tab=code`)
await page.getByText('README.md').first().waitFor({ timeout: 15000 })
await page.waitForTimeout(600)
await assertCleanState('Repo Overview')
const path1 = resolve(SITE_PRODUCT_DIR, 'forge-home.png')
await page.screenshot({ path: path1, type: 'png' })
writeFileSync(resolve(SITE_PRODUCT_DIR, 'repo-overview.png'), readFileSync(path1))
console.log(`Saved ${path1}`)

// View 2: Agent Console with Live Lease
console.log('2/5: Capturing agent console with live lease...')
await safeGoto(`http://127.0.0.1:${port}/?tab=agents`)
await page.getByText('AI Agent Telemetry').first().waitFor({ timeout: 15000 })
// Switch to Capability Leases tab
const leasesTabButton = page.getByRole('button', { name: /capability leases/i }).first()
if (await leasesTabButton.isVisible()) {
  await leasesTabButton.click()
  await page.waitForTimeout(400)
}
await page.getByText('Issued Agent Capability Leases').first().waitFor({ timeout: 10000 })
await page.waitForTimeout(600)
await assertCleanState('Agent Console')
const path2 = resolve(SITE_PRODUCT_DIR, 'forge-agents.png')
await page.screenshot({ path: path2, type: 'png' })
writeFileSync(resolve(SITE_PRODUCT_DIR, 'agent-console.png'), readFileSync(path2))
console.log(`Saved ${path2}`)

// View 3: PR with Checks and Merge Gate
console.log('3/5: Capturing PR with checks and merge gate...')
await safeGoto(`http://127.0.0.1:${port}/?tab=prs`)
const prRow = page.getByText('Add Forge capture note').first()
await prRow.waitFor({ timeout: 15000 })
await prRow.click()
await page.getByText('All checks have passed').first().waitFor({ timeout: 15000 })
await page.waitForTimeout(600)
await assertCleanState('Pull Request View')
const path3 = resolve(SITE_PRODUCT_DIR, 'forge-pr.png')
await page.screenshot({ path: path3, type: 'png' })
writeFileSync(resolve(SITE_PRODUCT_DIR, 'pr-checks.png'), readFileSync(path3))
writeFileSync(resolve(SITE_PUBLIC_DIR, 'forge-pr.png'), readFileSync(path3))
console.log(`Saved ${path3}`)

// View 4: Audit Log
console.log('4/5: Capturing audit log...')
await safeGoto(`http://127.0.0.1:${port}/?tab=admin`)
const auditNavButton = page.getByRole('button', { name: /audit log & security/i }).first()
await auditNavButton.waitFor({ timeout: 15000 })
await auditNavButton.click()
await page.waitForSelector('table', { timeout: 15000 })
await page.waitForTimeout(600)
await assertCleanState('Audit Log View')
const path4 = resolve(SITE_PRODUCT_DIR, 'forge-audit.png')
await page.screenshot({ path: path4, type: 'png' })
writeFileSync(resolve(SITE_PRODUCT_DIR, 'audit-log.png'), readFileSync(path4))
console.log(`Saved ${path4}`)

// View 5: MCP Configuration
console.log('5/5: Capturing MCP config...')
await safeGoto(`http://127.0.0.1:${port}/?tab=docs&article=quickstart`)
const mcpTabButton = page.getByRole('button', { name: /Claude Code \/ MCP Config/i }).first()
if (await mcpTabButton.isVisible()) {
  await mcpTabButton.click()
  await page.waitForTimeout(400)
}
const mcpConfigSection = page.getByText('Claude Code / MCP Config').first()
await mcpConfigSection.waitFor({ timeout: 15000 })
await mcpConfigSection.scrollIntoViewIfNeeded()
await page.waitForTimeout(600)
await assertCleanState('MCP Config View')
const path5 = resolve(SITE_PRODUCT_DIR, 'forge-mcp.png')
await page.screenshot({ path: path5, type: 'png' })
writeFileSync(resolve(SITE_PRODUCT_DIR, 'mcp-config.png'), readFileSync(path5))
console.log(`Saved ${path5}`)

await browser.close()

console.log('All 5 screenshots captured successfully!')
