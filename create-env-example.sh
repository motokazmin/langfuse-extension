#!/bin/bash

# Создание .env.example для cursor-langfuse-ext

set -e

echo "📝 Создаю .env.example для cursor-langfuse-ext..."

cat > cursor-langfuse-ext/.env.example << 'EOF'
# Langfuse API Configuration
# Получите ключи на https://cloud.langfuse.com или вашем self-hosted instance

# Public API Key (начинается с pk-lf-)
LANGFUSE_PUBLIC_KEY=pk-lf-your-public-key-here

# Secret API Key (начинается с sk-lf-)
LANGFUSE_SECRET_KEY=sk-lf-your-secret-key-here

# Langfuse Host URL
# Для cloud: https://cloud.langfuse.com
# Для self-hosted: ваш URL
LANGFUSE_BASE_URL=https://cloud.langfuse.com

# Инструкция:
# 1. Скопируйте этот файл: cp .env.example .env
# 2. Замените значения на ваши реальные ключи
# 3. НЕ коммитьте .env в Git!
EOF

echo "✅ Файл создан: cursor-langfuse-ext/.env.example"
cat cursor-langfuse-ext/.env.example
