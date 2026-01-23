#!/bin/bash

# Setup Cursor Infrastructure
# Для проекта langfuse-extension

set -e  # Прервать при ошибке

echo "🚀 Создание инфраструктуры Cursor..."

# 1. Создать структуру папок .cursor
echo "📁 Создаю директории .cursor..."
mkdir -p .cursor/{rules,context,analysis,plans,experiments,logs,summary}

# 2. Создать .cursorignore
echo "📝 Создаю .cursorignore..."
cat > .cursorignore << 'EOF'
# Dependencies
node_modules/
vendor/
.pnp/
.pnp.js

# Build outputs
dist/
build/
*.min.js
*.bundle.js

# Logs
*.log
logs/

# Environment
.env
.env.*
!.env.example

# IDE
.vscode/
.idea/

# Git
.git/

# Testing
coverage/
.nyc_output/
*.test
__tests__/

# Cursor hooks (не трогаем)
cursor-langfuse-ext/

# OS
.DS_Store
Thumbs.db

# Temporary
*.tmp
*.temp
.cache/
EOF

# 3. Проверить результат
echo ""
echo "✅ Структура создана:"
ls -la .cursor/
echo ""
echo "✅ .cursorignore создан"
echo ""
echo "🎉 Готово! Инфраструктура Cursor настроена."
