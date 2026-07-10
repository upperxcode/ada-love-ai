#!/bin/bash
# ============================================================
# Limpa Workspaces, Workers e Chats do Ada Love
# Agents NÃO são apagados (são executadores de tarefas, não chats)
# Uso: ./scripts/reset-data.sh
# ============================================================

set -e

CONFIG_DIR="$HOME/.config/ada-love"
CONFIG_FILE="$CONFIG_DIR/ada_config.json"
DB_FILE="$CONFIG_DIR/ada_love.db"

echo "========================================="
echo "  Ada Love - Reset de Dados"
echo "========================================="
echo ""

# 1. Reset ada_config.json (mantém agents)
if [ -f "$CONFIG_FILE" ]; then
  echo "[1/2] Limpando $CONFIG_FILE ..."
  python3 -c "
import json, sys

with open('$CONFIG_FILE', 'r') as f:
    d = json.load(f)

# Limpar apenas workspaces, workers e sessões
# Agents são executadores de tarefas - NÃO são apagados
d['workspaces'] = []
d['workers'] = []
d['active_workspace_path'] = ''
d['active_workspace_index'] = 0

with open('$CONFIG_FILE', 'w') as f:
    json.dump(d, f, indent=2, ensure_ascii=False)

print('  ✓ workspaces removidos')
print('  ✓ workers removidos')
print('  ℹ agents mantidos')
"
  echo ""
else
  echo "[1/2] $CONFIG_FILE não existe, pulando..."
  echo ""
fi

# 2. Reset banco de dados
if [ -f "$DB_FILE" ]; then
  echo "[2/2] Limpando $DB_FILE ..."
  sqlite3 "$DB_FILE" <<'SQL'
DELETE FROM messages;
DELETE FROM sessions;
DELETE FROM workspaces;
VACUUM;
SQL
  echo "  ✓ mensagens removidas"
  echo "  ✓ sessões removidas"
  echo "  ✓ workspaces removidos"
  echo ""
else
  echo "[2/2] $DB_FILE não existe, pulando..."
  echo ""
fi

echo "========================================="
echo "  ✓ Dados limpos!"
echo "  Reinicie o aplicativo para continuar."
echo "========================================="
