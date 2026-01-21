# Langfuse Analyzer Backend

Go backend для AI-анализа трейсов из Langfuse. Получает trace ID, анализирует через LLM, возвращает структурированные рекомендации.

---

## 🎯 Что он делает?

Backend — это мост между Chrome Extension и AI провайдером:

```
Chrome Extension → Backend → Langfuse API → AI Provider → Backend → Extension
```

**Основной workflow:**
1. Принимает HTTP запрос с trace ID от расширения
2. Запрашивает полный трейс из Langfuse API
3. Отправляет на AI провайдер (OpenRouter или Ollama)
4. Парсирует и валидирует результат
5. Возвращает структурированный JSON

**Детектирует 5 типов ситуаций:**
- 🐌 **PERFORMANCE_BOTTLENECK** — операции занимают >70% времени
- 💸 **HIGH_COST** — избыточные затраты (>$0.20 или >5000 токенов)
- 🔄 **LOGICAL_LOOP** — повторение операции >3 раз
- ❌ **ERROR** — ошибки выполнения (exit code != 0)
- ✅ **NONE** — всё в порядке

---

## 📋 Требования

- Go 1.21 или выше
- Langfuse credentials (public + secret key)
- AI Provider (один из):
  - OpenRouter API key (рекомендуется)
  - Ollama установленный локально

---

## 🚀 Установка

### Шаг 1: Настройте переменные окружения

```bash
cd ai-back
cp .env.example .env
```

Отредактируйте `.env`:

```env
# Langfuse API
LANGFUSE_PUBLIC_KEY=pk-lf-...
LANGFUSE_SECRET_KEY=sk-lf-...
LANGFUSE_BASEURL=https://cloud.langfuse.com

# AI Provider (выберите один)
AI_PROVIDER=openrouter          # или "ollama"

# OpenRouter (если AI_PROVIDER=openrouter)
OPENROUTER_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=anthropic/claude-3.5-sonnet

# Ollama (если AI_PROVIDER=ollama)
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=llama3.1:8b

# Server
PORT=8080
ALLOWED_ORIGINS=https://cloud.langfuse.com,chrome-extension://YOUR_EXTENSION_ID

# Optional
LOG_LEVEL=info
ENABLE_CORS=true
```

### Шаг 2: Получите API ключи

#### Langfuse

1. Откройте https://cloud.langfuse.com
2. Settings → API Keys → Create new
3. Скопируйте оба ключа

#### OpenRouter (рекомендуется)

1. Зарегистрируйтесь на https://openrouter.ai
2. Settings → API Keys → Create
3. Пополните баланс ($5 хватит на 500+ анализов)

#### Ollama (альтернатива)

```bash
# Установите Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Скачайте модель
ollama pull llama3.1:8b

# Проверьте работу
ollama run llama3.1:8b "Hello"
```

### Шаг 3: Настройте CORS

**⚠️ Критически важно для работы с Chrome Extension**

Backend должен разрешать запросы от расширения. Укажите ID вашего расширения:

```env
ALLOWED_ORIGINS=https://cloud.langfuse.com,chrome-extension://YOUR_EXTENSION_ID
```

**Как узнать ID расширения:**
1. Откройте `chrome://extensions/`
2. Найдите "Chrome Extension для Langfuse"
3. Скопируйте строку **ID:** из карточки

### Шаг 4: Запустите сервер

```bash
go run main.go
```

Ожидаемый вывод:
```
[GIN] Server listening on :8080
[INFO] AI Provider: openrouter
[INFO] Model: anthropic/claude-3.5-sonnet
[INFO] CORS enabled for: https://cloud.langfuse.com, chrome-extension://...
```

---

## ✅ Проверка работы

### Тест 1: Health Check

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "ai_provider": "openrouter",
  "model": "anthropic/claude-3.5-sonnet"
}
```

### Тест 2: Анализ трейса

Найдите любой trace ID в Langfuse:

```bash
curl "http://localhost:8080/analyze?traceId=YOUR_TRACE_ID"
```

Ожидаемый ответ (через 2-4 секунды):
```json
{
  "analysisSummary": {
    "overallStatus": "HEALTHY",
    "keyFinding": "Трейс выполнен успешно"
  },
  "detailedAnalysis": {
    "anomalyType": "NONE",
    "description": "Операции выполнены в разумное время",
    "rootCause": "Нет узких мест",
    "recommendation": "Продолжайте в том же духе"
  },
  "metadata": {
    "traceId": "YOUR_TRACE_ID",
    "analyzedAt": "2026-01-21T15:30:45Z",
    "processingTime": 2.3,
    "model": "anthropic/claude-3.5-sonnet",
    "provider": "openrouter"
  }
}
```

---

## 🎨 AI Провайдеры

### OpenRouter (рекомендуется)

**Преимущества:**
- Единый API для всех моделей (Claude, GPT-4, Gemini)
- Высокая надёжность и скорость
- Pay-as-you-go (платите только за использование)
- Выше rate limits чем у прямых API

**Недостатки:**
- Платно (~$0.001-0.02 за анализ)
- Данные отправляются на внешний сервис

**Рекомендуемые модели:**

| Модель | Скорость | Цена/анализ | Качество |
|--------|----------|-------------|----------|
| `google/gemini-2.0-flash-exp` | 1-2s | $0.001 | ⭐⭐⭐ |
| `meta-llama/llama-3.1-70b` | 3-5s | $0.005 | ⭐⭐⭐⭐ |
| `openai/gpt-4o` | 2-4s | $0.015 | ⭐⭐⭐⭐ |
| `anthropic/claude-3.5-sonnet` | 2-3s | $0.020 | ⭐⭐⭐⭐⭐ |

**Настройка:**
```env
AI_PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=google/gemini-2.0-flash-exp  # самая быстрая и дешёвая
```

---

### Ollama (для self-hosted)

**Преимущества:**
- Полностью бесплатно
- Работает локально (данные не покидают сервер)
- Нет rate limits
- Отлично для development и testing

**Недостатки:**
- Требует мощное железо (16GB+ RAM)
- Медленнее (5-10s на анализ)
- Качество ниже чем у Claude/GPT-4

**Рекомендуемые модели:**

| Модель | RAM | Скорость | Качество |
|--------|-----|----------|----------|
| `llama3.1:8b` | 8GB | 5-8s | ⭐⭐⭐ |
| `qwen2.5:14b` | 16GB | 8-12s | ⭐⭐⭐⭐ |
| `llama3.1:70b` | 48GB | 15-20s | ⭐⭐⭐⭐⭐ |

**Настройка:**
```env
AI_PROVIDER=ollama
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=llama3.1:8b
```

**Проверка:**
```bash
# Проверка что Ollama запущен
curl http://localhost:11434/api/version

# Список доступных моделей
ollama list
```

**Рекомендация:** Для production используйте OpenRouter, для development/self-hosted — Ollama.

---

## 🔧 API Reference

### `GET /health`

Проверка работоспособности сервера.

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "ai_provider": "openrouter",
  "model": "anthropic/claude-3.5-sonnet"
}
```

---

### `GET /analyze?traceId={id}`

Анализ трейса из Langfuse.

**Parameters:**
- `traceId` (required) — ID трейса из Langfuse

**Success Response (200):**
```json
{
  "analysisSummary": {
    "overallStatus": "WARNING | HEALTHY | ERROR",
    "keyFinding": "Краткое описание в одном предложении"
  },
  "detailedAnalysis": {
    "anomalyType": "NONE | PERFORMANCE_BOTTLENECK | HIGH_COST | LOGICAL_LOOP | ERROR",
    "description": "Детальное описание проблемы",
    "rootCause": "Гипотеза о причине",
    "recommendation": "Конкретный actionable совет"
  },
  "metadata": {
    "traceId": "f7b61b34-...",
    "analyzedAt": "2026-01-21T15:30:45Z",
    "processingTime": 2.3,
    "model": "anthropic/claude-3.5-sonnet",
    "provider": "openrouter"
  }
}
```

**Error Responses:**

| Code | Причина | Пример |
|------|---------|---------|
| 400 | Missing traceId | `{"error": "traceId parameter required"}` |
| 404 | Trace not found | `{"error": "Trace not found in Langfuse"}` |
| 429 | Rate limit | `{"error": "Too many requests", "retryAfter": 60}` |
| 500 | Server error | `{"error": "Internal server error"}` |
| 502 | AI provider error | `{"error": "AI provider unavailable"}` |

---

## 🔄 Как происходит анализ

### Пошаговый процесс

**1. Получение запроса**
```http
GET /analyze?traceId=f7b61b34-3a44-4146-9bd8-9f60fd788831
```

**2. Получение трейса из Langfuse**
```go
// С retry логикой (3 попытки с exponential backoff)
trace, err := langfuseClient.GetTrace(ctx, traceId)
```

**3. Подготовка данных для AI**
```go
traceData := TraceData{
  ID: trace.ID,
  Name: trace.Name,
  Metadata: trace.Metadata,
  Observations: trace.Observations,
  Latency: trace.Latency,
  TotalCost: trace.TotalCost,
}
```

**4. Отправка на AI провайдер**

Backend использует системный промпт, который определяет формат анализа:

```
Ты — 'TraceDebugger', AI-аналитик для LLM-приложений.

Проанализируй JSON-трейс из Langfuse и дай структурированный отчет.

Критерии детекции:
- PERFORMANCE_BOTTLENECK: latency > 10s ИЛИ одна операция >70% времени
- HIGH_COST: totalCost > $0.20 ИЛИ >5000 токенов на простой запрос
- LOGICAL_LOOP: операция повторяется >3 раз с похожими результатами
- ERROR: любые ошибки (exit code != 0, exceptions)
- NONE: всё в порядке

Формат ответа: JSON (см. документацию)
```

**5. Парсинг и валидация**
```go
var result AnalysisResult
if err := json.Unmarshal(aiResponse, &result); err != nil {
  return handleParseError(err)
}

// Валидация обязательных полей
if result.AnalysisSummary.OverallStatus == "" {
  return errors.New("missing overallStatus")
}
```

**6. Возврат результата**

Backend возвращает валидированный JSON клиенту.

---

## 🐛 Troubleshooting

### Backend не запускается

**Ошибка:** `missing LANGFUSE_PUBLIC_KEY`

**Решение:**
```bash
# Проверьте .env
cat .env | grep LANGFUSE

# Убедитесь что файл в правильном месте
ls -la .env
```

---

**Ошибка:** `port 8080 already in use`

**Решение:**
```bash
# Linux/Mac
lsof -ti:8080 | xargs kill -9

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Или измените порт в .env
PORT=8081
```

---

### Langfuse API ошибки

**Ошибка:** `401 Unauthorized`

**Решение:**
1. Проверьте правильность ключей в `.env`
2. Убедитесь что ключи активны в Langfuse Settings
3. Проверьте `LANGFUSE_BASEURL` (должен быть без `/` в конце)

---

**Ошибка:** `404 Trace not found`

**Решение:**
- Используйте правильный trace ID (скопируйте из URL)
- Убедитесь что API ключи от того же Langfuse проекта где находится трейс

---

### AI Provider ошибки

**Ошибка (OpenRouter):** `429 Rate limit exceeded`

**Решение:**
1. Подождите указанное время в `retryAfter`
2. Или смените модель на более дешёвую (`gemini-2.0-flash`)

---

**Ошибка (OpenRouter):** `402 Insufficient credits`

**Решение:**
1. Пополните баланс на https://openrouter.ai
2. Или временно переключитесь на Ollama

---

**Ошибка (Ollama):** `Connection refused`

**Решение:**
```bash
# Проверьте статус
systemctl status ollama

# Запустите
systemctl start ollama

# Или вручную
ollama serve
```

---

### Медленный анализ (>30s)

**Причины:**
1. Ollama на слабом CPU/GPU
2. Большой трейс (>100 observations)
3. Медленная сеть до OpenRouter

**Решения:**
- Используйте модель `gemini-2.0-flash` (самая быстрая)
- Переключитесь на OpenRouter если используете Ollama
- Для больших трейсов: увеличьте timeout в клиенте

---

## 🔒 Безопасность

### CORS

По умолчанию backend разрешает запросы только от:
- Указанных в `ALLOWED_ORIGINS` доменов
- Chrome Extension с указанным ID

**Важно:** Никогда не используйте `*` в production:
```go
// ❌ НЕ ДЕЛАЙТЕ ТАК
w.Header().Set("Access-Control-Allow-Origin", "*")

// ✅ ПРАВИЛЬНО
w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
```

### Данные и приватность

**Что отправляется на AI провайдер:**
- Полный JSON трейса (промпты, ответы, метаданные, токены)

**Рекомендации:**
- **Personal проекты:** OpenRouter OK
- **Корпоративные:** Self-hosted Ollama
- **Sensitive данные:** Только Ollama

**Логирование:**
Можно включить логирование запросов для аудита:
```env
LOG_LEVEL=debug  # Логировать все запросы к AI
```

---

## 📚 Дополнительные ресурсы

- **Langfuse API:** https://langfuse.com/docs/api
- **OpenRouter Docs:** https://openrouter.ai/docs
- **Ollama Documentation:** https://ollama.com/
- **Gin Framework:** https://gin-gonic.com/docs/

---

**Готово!** Backend настроен. Теперь установите [Chrome Extension](../chrome-ext/) для использования.