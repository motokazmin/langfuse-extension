#!/bin/bash

# Правильная настройка Langfuse hooks для проекта

set -e

echo "🔧 Настройка Langfuse hooks..."

# 1. Копируем hooks из cursor-langfuse-ext в .cursor
if [ -d "cursor-langfuse-ext/.cursor/hooks" ]; then
    echo "📁 Копирую hooks в .cursor/hooks..."
    mkdir -p .cursor/hooks
    cp -r cursor-langfuse-ext/.cursor/hooks/* .cursor/hooks/
    echo "✅ Hooks скопированы"
else
    echo "⚠️  cursor-langfuse-ext/.cursor/hooks не найден"
    echo "Создаю базовую структуру hooks..."
    mkdir -p .cursor/hooks
fi

# 1.5. Копируем hooks.json если есть
if [ -f "cursor-langfuse-ext/hooks.json" ]; then
    echo "📁 Копирую hooks.json..."
    cp cursor-langfuse-ext/hooks.json .cursor/
    echo "✅ hooks.json скопирован"
elif [ -f "cursor-langfuse-ext/.cursor/hooks.json" ]; then
    echo "📁 Копирую hooks.json..."
    cp cursor-langfuse-ext/.cursor/hooks.json .cursor/
    echo "✅ hooks.json скопирован"
else
    echo "⚠️  hooks.json не найден"
fi

# 2. Создаем .env.example в корне проекта
echo "📝 Создаю .env.example в корне..."
cat > .env.example << 'EOF'
# Langfuse Configuration (используется для cursor hooks и ai-back)
LANGFUSE_PUBLIC_KEY=pk-lf-your-public-key
LANGFUSE_SECRET_KEY=sk-lf-your-secret-key
LANGFUSE_BASE_URL=https://cloud.langfuse.com

# OpenRouter Configuration (для ai-back/ - AI анализ трейсов)
OPENROUTER_API_KEY=sk-or-your-api-key
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1

# Ollama Configuration (альтернатива OpenRouter для локального AI)
OLLAMA_BASE_URL=http://localhost:11434

# Инструкция:
# 1. cp .env.example .env
# 2. Заполните реальные значения
# 3. НЕ коммитьте .env!
EOF

echo "✅ .env.example создан в корне проекта"

# 3. Проверяем что .env в .gitignore
if ! grep -q "^\.env$" .gitignore 2>/dev/null; then
    echo "📝 Добавляю .env в .gitignore..."
    echo ".env" >> .gitignore
fi

echo ""
echo "🎉 Готово!"
echo ""
echo "Следующие шаги:"
echo "1. cp .env.example .env"
echo "2. Заполните .env реальными ключами"
echo "3. Проверьте .cursor/hooks/langfuse-client.js"
echo "4. git add .cursor/hooks/ .env.example .gitignore"
echo ""