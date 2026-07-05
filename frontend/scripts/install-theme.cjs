#!/usr/bin/env node
// npm run theme:add <tweakcn-url>
// Installs a theme from tweakcn and adds it to the theme registry.

const https = require('https')
const fs = require('fs')
const path = require('path')

const url = process.argv[2]
if (!url) {
  console.error('Usage: npm run theme:add <tweakcn-url>')
  console.error('Example: npm run theme:add https://tweakcn.com/r/themes/xxx')
  process.exit(1)
}

// Extract theme ID from URL
const themeId = url.split('/').pop() || `theme-${Date.now()}`
const registryPath = path.join(__dirname, '..', 'src', 'themes', 'registry.json')
const targetCss = path.join(__dirname, '..', 'src', 'style.css')

// Parse CSS variables from shadcn output CSS
function parseVars(css, selector) {
  const regex = new RegExp(`${selector}\\s*\\{([^}]+)\\}`, 'i')
  const match = css.match(regex)
  if (!match) return {}
  const vars = {}
  for (const line of match[1].split(';')) {
    const parts = line.split(':')
    if (parts.length >= 2) {
      const key = parts[0].trim().replace(/^--/, '')
      const val = parts.slice(1).join(':').trim()
      if (key && val) vars[key] = val
    }
  }
  return vars
}

function fetchUrl(url) {
  return new Promise((resolve, reject) => {
    https.get(url, res => {
      let data = ''
      res.on('data', chunk => data += chunk)
      res.on('end', () => resolve(data))
    }).on('error', reject)
  })
}

async function main() {
  console.log(`[theme:add] Fetching theme from ${url}...`)

  // For tweakcn, we need to get the raw CSS
  // The URL pattern is https://tweakcn.com/r/themes/<id>
  // We fetch it and extract the CSS
  const rawUrl = url.startsWith('http') ? url : `https://tweakcn.com/r/themes/${url}`
  let css

  try {
    css = await fetchUrl(rawUrl)
  } catch (e) {
    console.error(`[theme:add] Failed to fetch: ${e.message}`)
    process.exit(1)
  }

  // Parse the CSS to extract :root and .dark variables
  const light = parseVars(css, ':root')
  const dark = parseVars(css, '.dark')

  if (Object.keys(light).length === 0) {
    console.error('[theme:add] No CSS variables found in response')
    console.error('Make sure the URL points to a valid tweakcn theme CSS')
    process.exit(1)
  }

  // Add default button variables if missing from theme
  const btnLight = {
    'button-radius': '0.375rem',
    'button-font-weight': '500',
    'button-shadow': '0 1px 2px 0 oklch(0 0 0 / 0.05)',
    'button-hover-overlay': 'oklch(0 0 0 / 0.1)',
    'button-outline-hover-bg': 'var(--accent)'
  }
  const btnDark = {
    'button-radius': '0.375rem',
    'button-font-weight': '500',
    'button-shadow': '0 1px 2px 0 oklch(0 0 0 / 0.3)',
    'button-hover-overlay': 'oklch(1 1 1 / 0.1)',
    'button-outline-hover-bg': 'var(--accent)'
  }
  for (const [k, v] of Object.entries(btnLight)) { if (!light[k]) light[k] = v }
  for (const [k, v] of Object.entries(btnDark)) { if (!dark[k]) dark[k] = v }

  // Read existing registry
  let registry = []
  try {
    registry = JSON.parse(fs.readFileSync(registryPath, 'utf-8'))
  } catch {}

  // Check for duplicate
  const existing = registry.find(t => t.id === themeId)
  if (existing) {
    console.log(`[theme:add] Theme "${themeId}" already exists, updating...`)
  }

  const name = process.argv[3] || themeId
  const themeEntry = {
    id: themeId,
    name: name,
    author: 'tweakcn',
    url: url,
    light,
    dark
  }

  if (existing) {
    Object.assign(existing, themeEntry)
  } else {
    registry.push(themeEntry)
  }

  fs.writeFileSync(registryPath, JSON.stringify(registry, null, 2))
  console.log(`[theme:add] ✅ Theme "${name}" installed (${Object.keys(light).length} light, ${Object.keys(dark).length} dark vars)`)
  console.log(`[theme:add] Registry: ${registryPath}`)
}

main().catch(err => {
  console.error('[theme:add] Error:', err.message)
  process.exit(1)
})
