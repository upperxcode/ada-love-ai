#!/usr/bin/env node
// npm run icon-theme:add <package-name> "Theme Name" <icon-mapping-json>
// Example: npm run icon-theme:add "fi" "Feather Icons" '{"Plus":"FiPlus","Save":"FiSave"}'

const fs = require('fs')
const path = require('path')

const [, , pkg, name, mappingJson] = process.argv

if (!pkg || !name || !mappingJson) {
  console.error('Usage: npm run icon-theme:add <package-name> "Theme Name" \'{"IconName":"ExportName"}\'')
  console.error('Example: npm run icon-theme:add "fi" "Feather Icons" \'{"Plus":"FiPlus","Save":"FiSave"}\'')
  process.exit(1)
}

let mapping
try {
  mapping = JSON.parse(mappingJson)
} catch {
  console.error('[icon-theme] Invalid JSON mapping')
  process.exit(1)
}

const id = pkg
const registryPath = path.join(__dirname, '..', 'src', 'themes', 'icon-sets.json')
const registry = JSON.parse(fs.readFileSync(registryPath, 'utf-8'))

const existing = registry.find(t => t.id === id)
const entry = { id, name, pkg, mapping }

if (existing) {
  Object.assign(existing, entry)
  console.log(`[icon-theme] Updated "${name}"`)
} else {
  registry.push(entry)
  console.log(`[icon-theme] Installed "${name}"`)
}

fs.writeFileSync(registryPath, JSON.stringify(registry, null, 2))
console.log(`[icon-theme] Registry: ${registryPath}`)