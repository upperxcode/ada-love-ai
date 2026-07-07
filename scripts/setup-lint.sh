#!/usr/bin/env bash
set -euo pipefail

# setup-lint.sh — Instala e configura ESLint + Prettier no projeto
# Uso: ./scripts/setup-lint.sh

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo ""
echo "=========================================="
echo "  Setup ESLint + Prettier"
echo "=========================================="
echo ""

# --- 1. Verificar Node.js ---
if ! command -v node &>/dev/null; then
    err "Node.js não encontrado. Instale: https://nodejs.org"
    exit 1
fi
log "Node.js $(node --version) encontrado"

if ! command -v npm &>/dev/null; then
    err "npm não encontrado."
    exit 1
fi
log "npm $(npm --version) encontrado"

# --- 2. Instalar dependências ---
echo ""
echo "Instalando dependências..."
cd "$PROJECT_ROOT"

DEPS=(
    "prettier"
    "eslint@^9.0.0"
    "typescript"
    "@typescript-eslint/parser"
    "@typescript-eslint/eslint-plugin"
    "eslint-config-prettier"
)

npm install --save-dev "${DEPS[@]}" 2>&1 | tail -3
log "Dependências instaladas"

# --- 3. Criar .prettierrc ---
if [ -f "$PROJECT_ROOT/.prettierrc" ]; then
    warn ".prettierrc já existe, pulando"
else
    cat > "$PROJECT_ROOT/.prettierrc" <<'EOF'
{
  "semi": true,
  "trailingComma": "all",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2
}
EOF
    log ".prettierrc criado"
fi

# --- 4. Criar .prettierignore ---
if [ -f "$PROJECT_ROOT/.prettierignore" ]; then
    warn ".prettierignore já existe, pulando"
else
    cat > "$PROJECT_ROOT/.prettierignore" <<'EOF'
node_modules
dist
wailsjs
*.js
!.eslintrc.js
!eslint.config.mjs
EOF
    log ".prettierignore criado"
fi

# --- 5. Criar eslint.config.mjs ---
if [ -f "$PROJECT_ROOT/eslint.config.mjs" ]; then
    warn "eslint.config.mjs já existe, pulando"
else
    cat > "$PROJECT_ROOT/eslint.config.mjs" <<'EOF'
import tsParser from '@typescript-eslint/parser';
import tsPlugin from '@typescript-eslint/eslint-plugin';
import prettierConfig from 'eslint-config-prettier';

export default [
  {
    files: ['frontend/src/**/*.ts', 'frontend/src/**/*.tsx'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 2020,
        sourceType: 'module',
        ecmaFeatures: { jsx: true },
      },
    },
    plugins: {
      '@typescript-eslint': tsPlugin,
    },
    rules: {
      ...tsPlugin.configs.recommended.rules,
      '@typescript-eslint/no-unused-vars': 'warn',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-namespace': 'off',
    },
  },
  prettierConfig,
  {
    ignores: ['**/node_modules/**', '**/dist/**', '**/wailsjs/**'],
  },
];
EOF
    log "eslint.config.mjs criado"
fi

# --- 6. Adicionar scripts ao frontend/package.json ---
FRONTEND_PKG="$PROJECT_ROOT/frontend/package.json"
if [ -f "$FRONTEND_PKG" ]; then
    if grep -q '"format"' "$FRONTEND_PKG"; then
        warn "Scripts já existem em frontend/package.json, pulando"
    else
        # Adiciona scripts usando node para evitar erros de JSON
        node -e "
const fs = require('fs');
const pkg = JSON.parse(fs.readFileSync('$FRONTEND_PKG', 'utf8'));
pkg.scripts = pkg.scripts || {};
pkg.scripts.format = 'prettier --write \"src/**/*.ts\" \"src/**/*.tsx\"';
pkg.scripts.lint = 'eslint src/';
pkg.scripts['lint:fix'] = 'eslint src/ --fix';
fs.writeFileSync('$FRONTEND_PKG', JSON.stringify(pkg, null, 2) + '\n');
"
        log "Scripts adicionados em frontend/package.json"
    fi
else
    warn "frontend/package.json não encontrado, pulando scripts"
fi

# --- 7. Verificar instalação ---
echo ""
echo "Verificando instalação..."

cd "$PROJECT_ROOT"

if npx prettier --version &>/dev/null; then
    log "Prettier $(npx prettier --version) OK"
else
    err "Prettier falhou"
fi

if npx eslint --version &>/dev/null; then
    log "ESLint $(npx eslint --version) OK"
else
    err "ESLint falhou"
fi

# --- 8. Resumo ---
echo ""
echo "=========================================="
echo "  Setup concluído!"
echo "=========================================="
echo ""
echo "Arquivos criados/atualizados:"
echo "  - .prettierrc"
echo "  - .prettierignore"
echo "  - eslint.config.mjs"
echo "  - frontend/package.json (scripts)"
echo "  - package.json (devDependencies)"
echo ""
echo "Comandos disponíveis:"
echo "  cd frontend && npm run format     # Formata código"
echo "  cd frontend && npm run lint       # Verifica erros"
echo "  cd frontend && npm run lint:fix   # Corrige automaticamente"
echo ""
