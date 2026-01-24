package repository_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"langfuse-analyzer-backend/internal/repository"
)

// TestLangfuseRepository_GetTrace_Success тестирует успешное получение трейса
func TestLangfuseRepository_GetTrace_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/public/traces/test-trace-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Проверяем Basic Auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "pk-test", username)
		assert.Equal(t, "sk-test", password)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":   "test-trace-123",
			"name": "test trace",
			"spans": []map[string]interface{}{
				{
					"id":    "span-1",
					"name":  "span-name",
					"input": "test input",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace-123")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, "test-trace-123", trace["id"])
	assert.Equal(t, "test trace", trace["name"])
	spans := trace["spans"].([]interface{})
	assert.Len(t, spans, 1)
}

// TestLangfuseRepository_GetTrace_Retry тестирует логику повтора при ошибке сервера
func TestLangfuseRepository_GetTrace_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{"id": "test-trace-123"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace-123")

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "Should retry 3 times")
	assert.NotNil(t, trace)
	assert.Equal(t, "test-trace-123", trace["id"])
}

// TestLangfuseRepository_GetTrace_AllRetriesFail тестирует все попытки неудачны
func TestLangfuseRepository_GetTrace_AllRetriesFail(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace-123")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Equal(t, 3, attempts, "Should attempt 3 times")
	assert.Contains(t, err.Error(), "test-trace-123")
	assert.Contains(t, err.Error(), "after 3 attempts")
}

// TestLangfuseRepository_GetTrace_NotFound тестирует 404 ошибку
func TestLangfuseRepository_GetTrace_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "trace not found"}`))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "nonexistent")
}

// TestLangfuseRepository_GetTrace_Unauthorized тестирует 401 ошибку
func TestLangfuseRepository_GetTrace_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("wrong-key", "wrong-secret", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "401")
}

// TestLangfuseRepository_GetTrace_InvalidJSON тестирует некорректный JSON ответ
func TestLangfuseRepository_GetTrace_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json {]`))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "decode")
	assert.Contains(t, err.Error(), "test-trace")
}

// TestLangfuseRepository_GetTrace_EmptyResponse тестирует пустой JSON объект
func TestLangfuseRepository_GetTrace_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, 0, len(trace))
}

// TestLangfuseRepository_GetTrace_LargeResponse тестирует большой ответ
func TestLangfuseRepository_GetTrace_LargeResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":   "large-trace",
			"data": make([]string, 1000),
		}
		for i := 0; i < 1000; i++ {
			response["data"].([]string)[i] = "large data item"
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "large-trace")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, "large-trace", trace["id"])
	assert.Len(t, trace["data"].([]interface{}), 1000)
}

// TestLangfuseRepository_GetTrace_ServerError500 тестирует 500 ошибку сервера с retry
func TestLangfuseRepository_GetTrace_ServerError500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Equal(t, 3, attempts)
	assert.Contains(t, err.Error(), "500")
}

// TestLangfuseRepository_GetTrace_ServiceUnavailable тестирует 503 ошибку
func TestLangfuseRepository_GetTrace_ServiceUnavailable(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Unavailable"))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Equal(t, 3, attempts)
	assert.Contains(t, err.Error(), "503")
}

// TestLangfuseRepository_GetTrace_ContextCancellation тестирует отмену контекста
func TestLangfuseRepository_GetTrace_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{"id": "test-trace"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(ctx, "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	// Ошибка содержит информацию о трейсе
	assert.Contains(t, err.Error(), "test-trace")
}

// TestLangfuseRepository_GetTrace_SuccessAfterRetry тестирует успех после первой неудачи
func TestLangfuseRepository_GetTrace_SuccessAfterRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Bad Gateway"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":    "retry-trace",
			"retry": attempts,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "retry-trace")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, "retry-trace", trace["id"])
	assert.Equal(t, float64(2), trace["retry"])
	assert.Equal(t, 2, attempts)
}

// TestLangfuseRepository_GetTrace_MultipleSpans тестирует ответ с несколькими спанами
func TestLangfuseRepository_GetTrace_MultipleSpans(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id": "multi-span-trace",
			"spans": []map[string]interface{}{
				{"id": "span-1", "name": "first"},
				{"id": "span-2", "name": "second"},
				{"id": "span-3", "name": "third"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "multi-span-trace")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	spans := trace["spans"].([]interface{})
	assert.Len(t, spans, 3)

	firstSpan := spans[0].(map[string]interface{})
	assert.Equal(t, "span-1", firstSpan["id"])
	assert.Equal(t, "first", firstSpan["name"])
}

// TestLangfuseRepository_GetTrace_TraceLevelFields тестирует различные поля трейса
func TestLangfuseRepository_GetTrace_TraceLevelFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":           "complete-trace",
			"name":         "my-trace",
			"userId":       "user-123",
			"sessionId":    "session-456",
			"timestamp":    "2024-01-20T10:30:00Z",
			"duration":     1234,
			"status":       "success",
			"tags":         []string{"prod", "important"},
			"metadata":     map[string]interface{}{"version": "1.0"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "complete-trace")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, "complete-trace", trace["id"])
	assert.Equal(t, "my-trace", trace["name"])
	assert.Equal(t, "user-123", trace["userId"])
	assert.Equal(t, "session-456", trace["sessionId"])
	assert.Equal(t, float64(1234), trace["duration"])
	assert.Equal(t, "success", trace["status"])

	tags := trace["tags"].([]interface{})
	assert.Len(t, tags, 2)

	metadata := trace["metadata"].(map[string]interface{})
	assert.Equal(t, "1.0", metadata["version"])
}

// TestLangfuseRepository_GetTrace_URLEncoding тестирует корректное формирование URL с traceID
func TestLangfuseRepository_GetTrace_URLEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/public/traces/trace-with-special-chars-123", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{"id": "trace-with-special-chars-123"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "trace-with-special-chars-123")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
}

// TestLangfuseRepository_GetTrace_EmptyTraceID тестирует пустой traceID
func TestLangfuseRepository_GetTrace_EmptyTraceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/public/traces/", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "400")
}

// TestLangfuseRepository_GetTrace_ConnRefused тестирует отказ в соединении
func TestLangfuseRepository_GetTrace_ConnRefused(t *testing.T) {
	// Используем несуществующий адрес
	repo := repository.NewLangfuseRepository("pk-test", "sk-test", "http://127.0.0.1:1")

	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "test-trace")
	assert.Contains(t, err.Error(), "after 3 attempts")
}

// TestLangfuseRepository_GetTrace_NestedStructure тестирует вложенные структуры в ответе
func TestLangfuseRepository_GetTrace_NestedStructure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id": "nested-trace",
			"spans": []map[string]interface{}{
				{
					"id": "span-1",
					"events": []map[string]interface{}{
						{"type": "start", "timestamp": "2024-01-20T10:30:00Z"},
						{"type": "end", "timestamp": "2024-01-20T10:30:01Z"},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "nested-trace")

	assert.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, "nested-trace", trace["id"])

	spans := trace["spans"].([]interface{})
	assert.Len(t, spans, 1)

	firstSpan := spans[0].(map[string]interface{})
	events := firstSpan["events"].([]interface{})
	assert.Len(t, events, 2)

	firstEvent := events[0].(map[string]interface{})
	assert.Equal(t, "start", firstEvent["type"])
}

// TestLangfuseRepository_GetTrace_BasicAuthPresent проверяет наличие Basic Auth в запросе
func TestLangfuseRepository_GetTrace_BasicAuthPresent(t *testing.T) {
	authChecked := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		authChecked = true
		assert.True(t, ok, "Basic Auth should be present")
		assert.Equal(t, "pk-test", username)
		assert.Equal(t, "sk-test", password)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{"id": "test-trace"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.NoError(t, err)
	assert.True(t, authChecked)
	assert.NotNil(t, trace)
}

// TestLangfuseRepository_NewRepository проверяет корректное создание репозитория
func TestLangfuseRepository_NewRepository(t *testing.T) {
	repo := repository.NewLangfuseRepository("pk-test", "sk-test", "http://localhost:3000")
	assert.NotNil(t, repo)

	// Проверяем, что это имплементирует интерфейс
	var _ repository.LangfuseRepository = repo
}

// TestLangfuseRepository_GetTrace_ForbiddenResponse тестирует 403 ошибку
func TestLangfuseRepository_GetTrace_ForbiddenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "access denied"}`))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), "403")
}

// TestLangfuseRepository_GetTrace_RateLimit тестирует 429 ошибку (rate limit)
func TestLangfuseRepository_GetTrace_RateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "too many requests"}`))
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	trace, err := repo.GetTrace(context.Background(), "test-trace")

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Equal(t, 3, attempts)
	assert.Contains(t, err.Error(), "429")
}

// TestLangfuseRepository_GetTrace_TraceIDInError проверяет, что traceID присутствует в сообщении об ошибке
func TestLangfuseRepository_GetTrace_TraceIDInError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)
	traceID := "specific-trace-id-12345"
	trace, err := repo.GetTrace(context.Background(), traceID)

	assert.Error(t, err)
	assert.Nil(t, trace)
	assert.Contains(t, err.Error(), traceID, "Error message should contain traceID for debugging")
}

// BenchmarkLangfuseRepository_GetTrace бенчмарк для измерения производительности
func BenchmarkLangfuseRepository_GetTrace(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":   "benchmark-trace",
			"name": "benchmark",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := repository.NewLangfuseRepository("pk-test", "sk-test", server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetTrace(context.Background(), "benchmark-trace")
	}
}
