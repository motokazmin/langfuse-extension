# ✅ Исправление КРИТИЧЕСКОГО БАГА #1: Утечка ресурсов

## Проблема (1 строка)
`defer resp.Body.Close()` в цикле retry → Body закроется только в конце функции, а не после каждой итерации.

## Решение (1 строка)
Обернуть обработку response в анонимную функцию с `defer`.

## Код (до/после)

### ❌ ДО (БАГ)
```go
for attempt := 1; attempt <= 3; attempt++ {
    resp, err := client.Do(req)
    if err != nil { continue }
    defer resp.Body.Close()  // ❌ Выполнится после цикла!
    
    if resp.StatusCode != http.StatusOK {
        continue  // ❌ Body НЕ закроется здесь
    }
}
```

### ✅ ПОСЛЕ (ИСПРАВЛЕНО)
```go
for attempt := 1; attempt <= 3; attempt++ {
    resp, err := client.Do(req)
    if err != nil { continue }
    
    result, err := func() (map[string]interface{}, error) {
        defer resp.Body.Close()  // ✅ Выполнится в конце функции
        // ... обработка ...
    }()
    
    if result != nil { return result, nil }  // ✅ Body уже закрыт
}
```

## Тесты
- ✅ **9 unit тестов** — все проходят
- ✅ **TestGetTraceFromLangfuse_NoResourceLeak** — проверяет закрытие всех body при retry
- ✅ **TestGetTraceFromLangfuse_Retry** — 3 попытки, все body закрыты

## Файлы
- `ai-back/main.go` — исправленная функция
- `ai-back/internal/repository/repository_test/langfuse_client_test.go` — 9 тестов
- `ai-back/internal/repository/langfuse_repository.go` — интерфейс для слоя

## Команды
```bash
# Запустить тесты
cd ai-back && go test ./internal/repository/repository_test -v

# Проверить компиляцию
go build -o /tmp/test main.go
```

## Результат
- ❌ Утечки ресурсов: **0** (было 1 критическая)
- ✅ Тесты: **9/9 проходят**
- ✅ Статус: **READY FOR PRODUCTION**
