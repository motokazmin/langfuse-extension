#!/bin/bash

# Анализ структуры проекта langfuse-extension (исправленный)
# Исключаем node_modules, dist, build, vendor

set -e

echo "🔍 Анализ структуры проекта (только исходный код)..."
echo ""

# Общее количество файлов
echo "📊 Общее количество исходных файлов:"
find . -type f \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" -o -name "*.py" \) \
  -not -path "*/node_modules/*" \
  -not -path "*/cursor-langfuse-ext/*" \
  -not -path "*/dist/*" \
  -not -path "*/build/*" \
  -not -path "*/vendor/*" \
  -not -path "*/.cache/*" | wc -l

echo ""
echo "📝 Размер кодовой базы по языкам:"
echo ""

# Go files
echo "=== Go files ==="
GO_LINES=$(find ai-back test-tasks -name "*.go" \
  -not -path "*/vendor/*" 2>/dev/null \
  -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
echo "${GO_LINES:-0} total lines in Go"
echo ""

# TypeScript files
echo "=== TypeScript files ==="
TS_LINES=$(find crome-ext -name "*.ts" -o -name "*.tsx" \
  -not -path "*/node_modules/*" \
  -not -path "*/dist/*" 2>/dev/null \
  -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
echo "${TS_LINES:-0} total lines in TypeScript"
echo ""

# JavaScript files (только src, не dist)
echo "=== JavaScript files ==="
JS_LINES=$(find crome-ext/src -name "*.js" -o -name "*.jsx" 2>/dev/null \
  -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
echo "${JS_LINES:-0} total lines in JavaScript"
echo ""

# Python files
echo "=== Python files ==="
PY_LINES=$(find test-tasks -name "*.py" 2>/dev/null \
  -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
echo "${PY_LINES:-0} total lines in Python"
echo ""

# Список файлов по компонентам
echo "📁 Исходные файлы по компонентам:"
echo ""

echo "--- ai-back/ ---"
find ai-back -name "*.go" -not -path "*/vendor/*" 2>/dev/null | while read f; do
  lines=$(wc -l "$f" 2>/dev/null | awk '{print $1}')
  echo "$f: $lines lines"
done
echo ""

echo "--- crome-ext/src/ ---"
find crome-ext/src -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" 2>/dev/null | while read f; do
  lines=$(wc -l "$f" 2>/dev/null | awk '{print $1}')
  echo "$f: $lines lines"
done
echo ""

echo "--- test-tasks/ ---"
find test-tasks -name "*.py" 2>/dev/null | while read f; do
  lines=$(wc -l "$f" 2>/dev/null | awk '{print $1}')
  echo "$f: $lines lines"
done
echo ""

# Итоговая статистика
echo "📊 Итоговая статистика:"
TOTAL=$((${GO_LINES:-0} + ${TS_LINES:-0} + ${JS_LINES:-0} + ${PY_LINES:-0}))
echo "Всего строк исходного кода: $TOTAL"
echo ""
echo "✅ Анализ завершен"