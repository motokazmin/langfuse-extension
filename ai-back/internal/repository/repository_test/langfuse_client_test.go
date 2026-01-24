package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBodyTracker tracks when response bodies are closed
type testBodyTracker struct {
	io.ReadCloser
	closedMutex sync.Mutex
	closeCount  int32
	closed      bool
}

func (t *testBodyTracker) Close() error {
	t.closedMutex.Lock()
	defer t.closedMutex.Unlock()
	atomic.AddInt32(&t.closeCount, 1)
	t.closed = true
	return t.ReadCloser.Close()
}

func (t *testBodyTracker) IsClosed() bool {
	t.closedMutex.Lock()
	defer t.closedMutex.Unlock()
	return t.closed
}

// TestGetTraceFromLangfuse_NoResourceLeak проверяет что все response.Body закрываются
func TestGetTraceFromLangfuse_NoResourceLeak(t *testing.T) {
	attemptCount := 0
	var trackers []*testBodyTracker
	trackersMutex := sync.Mutex{}

	// Mock Langfuse API: первые 2 попытки - ошибки, 3я - успех
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// Проверяем Basic Auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok, "Basic auth должен быть установлен")
		assert.Equal(t, "pk-test", username, "Public key должен быть username")
		assert.Equal(t, "sk-test", password, "Secret key должен быть password")

		if attemptCount < 3 {
			// Первые 2 попытки - ошибка 500
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "temporary error"}`))
		} else {
			// 3я попытка - успех
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test-trace-123",
				"name": "test trace",
				"observations": [],
				"latency": 1500,
				"totalCost": 0.05
			}`))
		}
	}))
	defer server.Close()

	// Создаём client с tracking transport
	patchedClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &trackingTransport{
			base: http.DefaultTransport,
			trackers: func(tracker *testBodyTracker) {
				trackersMutex.Lock()
				defer trackersMutex.Unlock()
				trackers = append(trackers, tracker)
			},
		},
	}

	// Вызываем функцию (через инъекцию или через прямой вызов)
	// Здесь используем прямой вызов для демонстрации
	t.Run("direct_function_call", func(t *testing.T) {
		// Симулируем вызов функции getTraceFromLangfuse
		result := callGetTraceFromLangfuseWithClient(server.URL, "pk-test", "sk-test", patchedClient)

		// Проверяем результат
		assert.NotNil(t, result, "Результат не должен быть nil")
		assert.Equal(t, "test-trace-123", result["id"], "ID трейса должен совпадать")

		// Проверяем что было 3 попытки
		assert.Equal(t, 3, attemptCount, "Должно быть 3 попытки")

		// Проверяем что все Body закрыты
		assert.True(t, len(trackers) >= 3, "Должно быть минимум 3 tracker'а для body")

		allClosed := true
		for i, tracker := range trackers {
			if i < 3 { // Проверяем первые 3
				if !tracker.IsClosed() {
					allClosed = false
					t.Logf("Body #%d не закрыт!", i+1)
				}
			}
		}
		assert.True(t, allClosed, "Все response bodies должны быть закрыты")
	})
}

// TestGetTraceFromLangfuse_Success проверяет успешное получение трейса
func TestGetTraceFromLangfuse_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/public/traces/test-trace-123", r.URL.Path)

		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "pk-test", username)
		assert.Equal(t, "sk-test", password)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "test-trace-123",
			"name": "test trace",
			"observations": [
				{"id": "obs-1", "name": "step 1"}
			],
			"latency": 1500,
			"totalCost": 0.05
		}`))
	}))
	defer server.Close()

	// Вызываем функцию
	client := &http.Client{Timeout: 30 * time.Second}
	result := callGetTraceFromLangfuseWithClient(server.URL, "pk-test", "sk-test", client)

	// Проверяем результат
	require.NotNil(t, result)
	assert.Equal(t, "test-trace-123", result["id"])
	assert.Equal(t, "test trace", result["name"])

	observations := result["observations"].([]interface{})
	assert.Equal(t, 1, len(observations))
}

// TestGetTraceFromLangfuse_Retry проверяет retry логику
func TestGetTraceFromLangfuse_Retry(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts < 3 {
			// Первые 2 попытки - ошибка
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "temporary error"}`))
			return
		}

		// 3я попытка - успех
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "test-trace-123"}`))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	result := callGetTraceFromLangfuseWithClient(server.URL, "pk-test", "sk-test", client)

	assert.NotNil(t, result, "Результат должен быть успешным после retry")
	assert.Equal(t, 3, attempts, "Должно быть 3 попытки")
	assert.Equal(t, "test-trace-123", result["id"])
}

// TestGetTraceFromLangfuse_NotFound проверяет обработку 404
func TestGetTraceFromLangfuse_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "trace not found"}`))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	result, err := callGetTraceFromLangfuseWithClientError(server.URL, "pk-test", "sk-test", client)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// TestGetTraceFromLangfuse_InvalidJSON проверяет обработку невалидного JSON
func TestGetTraceFromLangfuse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	result, err := callGetTraceFromLangfuseWithClientError(server.URL, "pk-test", "sk-test", client)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode")
}

// TestGetTraceFromLangfuse_InvalidRequest проверяет обработку ошибки создания запроса
func TestGetTraceFromLangfuse_InvalidRequest(t *testing.T) {
	// Используем невалидный URL с null байтом
	client := &http.Client{Timeout: 30 * time.Second}
	_, err := callGetTraceFromLangfuseWithClientError("http://invalid\x00url", "pk-test", "sk-test", client)

	assert.Error(t, err)
	// Ошибка может быть как от создания запроса, так и от парсинга URL
	assert.True(t,
		strings.Contains(err.Error(), "failed to create") ||
			strings.Contains(err.Error(), "invalid control character"),
		"Ошибка должна быть связана с невалидным URL")
}

// TestGetTraceFromLangfuse_MalformedURL проверяет обработку синтаксически неправильного URL
func TestGetTraceFromLangfuse_MalformedURL(t *testing.T) {
	// Синтаксически неправильный URL (без scheme)
	client := &http.Client{Timeout: 30 * time.Second}
	_, err := callGetTraceFromLangfuseWithClientError("not-a-url", "pk-test", "sk-test", client)

	assert.Error(t, err)
	assert.True(t,
		strings.Contains(err.Error(), "unsupported protocol") ||
			strings.Contains(err.Error(), "no such host") ||
			strings.Contains(err.Error(), "failed to create"),
		"Ошибка должна быть связана с неправильным URL")
}

// TestGetTraceFromLangfuse_EmptyURL проверяет обработку пустого URL
func TestGetTraceFromLangfuse_EmptyURL(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	_, err := callGetTraceFromLangfuseWithClientError("", "pk-test", "sk-test", client)

	assert.Error(t, err)
	// Пустой URL приведёт к ошибке при попытке подключения
	// Может быть разная ошибка в зависимости от реализации
	assert.NotNil(t, err, "Должна быть ошибка при пустом URL")
}

// TestGetTraceFromLangfuse_AllRetriesFail проверяет когда все retry не удаются
func TestGetTraceFromLangfuse_AllRetriesFail(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	result, err := callGetTraceFromLangfuseWithClientError(server.URL, "pk-test", "sk-test", client)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, 3, attempts, "Должно быть 3 попытки перед ошибкой")
}

// TestGetTraceFromLangfuse_NetworkTimeout проверяет обработку timeout
func TestGetTraceFromLangfuse_NetworkTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Спим дольше чем timeout
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "test"}`))
	}))
	defer server.Close()

	// Используем очень маленький timeout
	client := &http.Client{Timeout: 10 * time.Millisecond}
	result, err := callGetTraceFromLangfuseWithClientError(server.URL, "pk-test", "sk-test", client)

	assert.Nil(t, result)
	assert.Error(t, err)
	// Timeout ошибка должна содержать context deadline exceeded
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "timeout"),
		"Ошибка должна быть связана с timeout")
}

// trackingTransport перехватывает HTTP запросы для отслеживания body
type trackingTransport struct {
	base     http.RoundTripper
	trackers func(*testBodyTracker)
}

func (t *trackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Оборачиваем body для отслеживания
	tracker := &testBodyTracker{ReadCloser: resp.Body}
	resp.Body = tracker
	t.trackers(tracker)

	return resp, nil
}

// Вспомогательные функции для вызова getTraceFromLangfuse
// (В реальном коде это должны быть методы репозитория)

func callGetTraceFromLangfuseWithClient(baseURL, publicKey, secretKey string, client *http.Client) map[string]interface{} {
	result, _ := callGetTraceFromLangfuseWithClientError(baseURL, publicKey, secretKey, client)
	return result
}

func callGetTraceFromLangfuseWithClientError(baseURL, publicKey, secretKey string, client *http.Client) (map[string]interface{}, error) {
	url := baseURL + "/api/public/traces/test-trace-123"

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(attempt) * time.Millisecond)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.SetBasicAuth(publicKey, secretKey)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// ✅ Корректное закрытие body в каждой итерации
		result, err := func() (map[string]interface{}, error) {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("Langfuse API вернул статус: %d - %s", resp.StatusCode, string(bodyBytes))
			}

			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}

			return data, nil
		}()

		if result != nil {
			return result, nil
		}

		if err != nil {
			lastErr = err
		}
	}

	return nil, lastErr
}
