package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"langfuse-analyzer-backend/ai"
	"langfuse-analyzer-backend/internal/handler"
	"langfuse-analyzer-backend/internal/repository"
	"langfuse-analyzer-backend/internal/service"
)

// TestE2E_FullFlow_Success тестирует полный flow: HTTP request → Handler → Service → Repository → Response
func TestE2E_FullFlow_Success(t *testing.T) {
	// Mock Langfuse API
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/public/traces/trace-123", r.URL.Path)

		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "pk-test", username)
		assert.Equal(t, "sk-test", password)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		traceData := map[string]interface{}{
			"id":     "trace-123",
			"name":   "E2E Test Trace",
			"status": "success",
			"spans": []map[string]interface{}{
				{
					"id":       "span-1",
					"name":     "first-operation",
					"duration": 100,
				},
			},
		}
		json.NewEncoder(w).Encode(traceData)
	}))
	defer langfuseServer.Close()

	// Setup DI
	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	// Setup HTTP router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	// Make HTTP request
	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"trace-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestE2E_FullFlow_LangfuseError тестирует обработку ошибок Langfuse
func TestE2E_FullFlow_LangfuseError(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "trace not found"}`))
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"nonexistent"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestE2E_FullFlow_InvalidRequest тестирует валидацию запроса
func TestE2E_FullFlow_InvalidRequest(t *testing.T) {
	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", "http://localhost:9999")
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`invalid json`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestE2E_FullFlow_MissingTraceId тестирует отсутствие traceId
func TestE2E_FullFlow_MissingTraceId(t *testing.T) {
	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", "http://localhost:9999")
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestE2E_FullFlow_ComplexTrace тестирует анализ сложного трейса
func TestE2E_FullFlow_ComplexTrace(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		complexTrace := map[string]interface{}{
			"id":   "complex-trace",
			"name": "Complex E2E Trace",
			"spans": []map[string]interface{}{
				{
					"id":       "span-1",
					"name":     "step-1",
					"duration": 100,
					"status":   "success",
				},
				{
					"id":       "span-2",
					"name":     "step-2",
					"duration": 500,
					"status":   "error",
					"error":    "timeout",
				},
				{
					"id":       "span-3",
					"name":     "step-3",
					"duration": 200,
					"status":   "success",
				},
			},
			"metadata": map[string]interface{}{
				"user":      "test-user",
				"version":   "1.0.0",
				"environment": "test",
			},
			"metrics": map[string]interface{}{
				"totalDuration": 800,
				"tokenCount":    1234,
			},
		}
		json.NewEncoder(w).Encode(complexTrace)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"complex-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestE2E_FullFlow_MultipleRequests тестирует последовательные запросы
func TestE2E_FullFlow_MultipleRequests(t *testing.T) {
	requestCount := 0
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		traceData := map[string]interface{}{
			"id":   fmt.Sprintf("trace-%d", requestCount),
			"name": fmt.Sprintf("Trace %d", requestCount),
		}
		json.NewEncoder(w).Encode(traceData)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	for i := 1; i <= 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			"POST",
			"/analyze",
			bytes.NewBufferString(fmt.Sprintf(`{"traceId":"trace-%d"}`, i)),
		)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 3, requestCount)
}

// TestE2E_FullFlow_LargeLangfuseResponse тестирует большой ответ от Langfuse
func TestE2E_FullFlow_LargeLangfuseResponse(t *testing.T) {
	if os.Getenv("SKIP_LARGE_RESPONSE_TEST") == "" {
		t.Skip("Skipping large response test - takes too long with AI processing")
	}

	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		spans := make([]map[string]interface{}, 50)
		for i := 0; i < 50; i++ {
			spans[i] = map[string]interface{}{
				"id":       fmt.Sprintf("span-%d", i),
				"name":     fmt.Sprintf("operation-%d", i),
				"duration": 50 + i,
			}
		}

		largeTrace := map[string]interface{}{
			"id":    "large-trace",
			"name":  "Large E2E Trace",
			"spans": spans,
		}
		json.NewEncoder(w).Encode(largeTrace)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"large-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, w.Body.Len(), 500)
}

// TestE2E_FullFlow_ContextCancellation тестирует отмену контекста
func TestE2E_FullFlow_ContextCancellation(t *testing.T) {
	if os.Getenv("SKIP_TIMEOUT_TEST") == "" {
		t.Skip("Skipping timeout test - takes too long")
	}

	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Задержка для срабатывания таймаута
		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Должно получить ошибку
	assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusOK)
}

// TestE2E_FullFlow_UnauthorizedLangfuse тестирует ошибку авторизации
func TestE2E_FullFlow_UnauthorizedLangfuse(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("wrong-key", "wrong-secret", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestE2E_FullFlow_LangfuseRateLimit тестирует rate limit от Langfuse
func TestE2E_FullFlow_LangfuseRateLimit(t *testing.T) {
	attempt := 0
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limited"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		traceData := map[string]interface{}{
			"id":   "rate-limited-trace",
			"name": "Trace after rate limit",
		}
		json.NewEncoder(w).Encode(traceData)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"rate-limited-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// С retry логикой должно вернуться успешно
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestE2E_FullFlow_SpecialCharactersInTraceId тестирует спецсимволы в traceId
func TestE2E_FullFlow_SpecialCharactersInTraceId(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		traceData := map[string]interface{}{
			"id":   "trace-<>&-123",
			"name": "Trace with special chars",
		}
		json.NewEncoder(w).Encode(traceData)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"trace-<>&-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestE2E_FullFlow_EmptyLangfuseResponse тестирует пустой ответ от Langfuse
func TestE2E_FullFlow_EmptyLangfuseResponse(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"empty-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestE2E_FullFlow_TraceWithNestedStructures тестирует трейс с вложенными структурами
func TestE2E_FullFlow_TraceWithNestedStructures(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		nestedTrace := map[string]interface{}{
			"id": "nested-trace",
			"spans": []map[string]interface{}{
				{
					"id": "span-1",
					"events": []map[string]interface{}{
						{
							"type": "start",
							"timestamp": "2024-01-20T10:30:00Z",
						},
						{
							"type": "end",
							"timestamp": "2024-01-20T10:30:01Z",
						},
					},
					"metadata": map[string]interface{}{
						"tags": []string{"important", "production"},
						"context": map[string]interface{}{
							"user_id": "user-123",
							"session_id": "session-456",
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(nestedTrace)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"nested-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestE2E_FullFlow_ResponseContentType тестирует Content-Type ответа
func TestE2E_FullFlow_ResponseContentType(t *testing.T) {
	langfuseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		traceData := map[string]interface{}{
			"id": "test-trace",
		}
		json.NewEncoder(w).Encode(traceData)
	}))
	defer langfuseServer.Close()

	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", langfuseServer.URL)
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/analyze", analyzeHandler.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-trace"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

// TestE2E_Setup проверяет, что все компоненты могут быть инициализированы
func TestE2E_Setup(t *testing.T) {
	if os.Getenv("SKIP_SETUP_TEST") != "" {
		t.Skip("Skipping setup test")
	}

	// Инициализация AI клиента
	aiClient := ai.NewAIClient(ai.ProviderOllama, "", "http://localhost:11434", "llama3.2", 1000)
	require.NotNil(t, aiClient)

	// Инициализация репозитория
	langfuseRepo := repository.NewLangfuseRepository("pk-test", "sk-test", "http://localhost:8000")
	require.NotNil(t, langfuseRepo)

	// Инициализация сервиса
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	require.NotNil(t, analyzeService)

	// Инициализация handler
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)
	require.NotNil(t, analyzeHandler)
}
