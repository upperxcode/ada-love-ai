@echo off
REM setup-lint.bat — Instala e configura ESLint + Prettier no projeto (Windows)
REM Uso: scripts\setup-lint.bat

echo.
echo ==========================================
echo   Setup ESLint + Prettier
echo ==========================================
echo.

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [X] Node.js nao encontrado. Instale: https://nodejs.org
    exit /b 1
)
echo [OK] Node.js encontrado

where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo [X] npm nao encontrado.
    exit /b 1
)
echo [OK] npm encontrado

echo.
echo Instalando dependencias...
cd /d "%~dp0.."

call npm install --save-dev prettier eslint@^9.0.0 typescript @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-config-prettier
if %errorlevel% neq 0 (
    echo [X] Falha ao instalar dependencias
    exit /b 1
)
echo [OK] Dependencias instaladas

REM .prettierrc
if exist ".prettierrc" (
    echo [!] .prettierrc ja existe, pulando
) else (
    echo {"semi": true, "trailingComma": "all", "singleQuote": true, "printWidth": 80, "tabWidth": 2} > .prettierrc
    echo [OK] .prettierrc criado
)

REM .prettierignore
if exist ".prettierignore" (
    echo [!] .prettierignore ja existe, pulando
) else (
    (
        echo node_modules
        echo dist
        echo wailsjs
        echo *.js
        echo !.eslintrc.js
        echo !eslint.config.mjs
    ) > .prettierignore
    echo [OK] .prettierignore criado
)

REM eslint.config.mjs
if exist "eslint.config.mjs" (
    echo [!] eslint.config.mjs ja existe, pulando
) else (
    (
        echo import tsParser from '@typescript-eslint/parser';
        echo import tsPlugin from '@typescript-eslint/eslint-plugin';
        echo import prettierConfig from 'eslint-config-prettier';
        echo.
        echo export default [
        echo   {
        echo     files: ['frontend/src/**/*.ts', 'frontend/src/**/*.tsx'],
        echo     languageOptions: {
        echo       parser: tsParser,
        echo       parserOptions: {
        echo         ecmaVersion: 2020,
        echo         sourceType: 'module',
        echo         ecmaFeatures: { jsx: true },
        echo       },
        echo     },
        echo     plugins: {
        echo       '@typescript-eslint': tsPlugin,
        echo     },
        echo     rules: {
        echo       ...tsPlugin.configs.recommended.rules,
        echo       '@typescript-eslint/no-unused-vars': 'warn',
        echo       '@typescript-eslint/no-explicit-any': 'warn',
        echo       '@typescript-eslint/no-namespace': 'off',
        echo     },
        echo   },
        echo   prettierConfig,
        echo   {
        echo     ignores: ['**/node_modules/**', '**/dist/**', '**/wailsjs/**'],
        echo   },
        echo ];
    ) > eslint.config.mjs
    echo [OK] eslint.config.mjs criado
)

echo.
echo ==========================================
echo   Setup concluido!
echo ==========================================
echo.
echo Comandos disponiveis:
echo   cd frontend ^&^& npm run format
echo   cd frontend ^&^& npm run lint
echo   cd frontend ^&^& npm run lint:fix
echo.
