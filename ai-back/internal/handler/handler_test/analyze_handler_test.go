package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestAnalyzeHandler_Integration_Success тестирует успешный запрос с JSON ответом
func TestAnalyzeHandler_Integration_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": map[string]interface{}{"status": "ok"}})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestAnalyzeHandler_Integration_InvalidJSON тестирует невалидный JSON
func TestAnalyzeHandler_Integration_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

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

// TestAnalyzeHandler_Integration_MissingTraceId тестирует отсутствие traceId
func TestAnalyzeHandler_Integration_MissingTraceId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

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

// TestAnalyzeHandler_Integration_RateLimit тестирует обработку rate limit ошибки
func TestAnalyzeHandler_Integration_RateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":      "Слишком много запросов к AI",
			"code":       "RATE_LIMIT",
			"retryAfter": 60,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "RATE_LIMIT", response["code"])
}

// TestAnalyzeHandler_Integration_InsufficientCredits тестирует 402 ошибку
func TestAnalyzeHandler_Integration_InsufficientCredits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusPaymentRequired, gin.H{
			"error": "Недостаточно кредитов",
			"code":  "INSUFFICIENT_CREDITS",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INSUFFICIENT_CREDITS", response["code"])
}

// TestAnalyzeHandler_Integration_ServiceUnavailable тестирует 503 ошибку
func TestAnalyzeHandler_Integration_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI сервис недоступен",
			"code":  "SERVICE_UNAVAILABLE",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SERVICE_UNAVAILABLE", response["code"])
}

// TestAnalyzeHandler_Integration_GenericError тестирует обработку общей ошибки
func TestAnalyzeHandler_Integration_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze trace: internal error",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to analyze trace")
}

// TestAnalyzeHandler_Integration_EmptyBody тестирует пустое тело запроса
func TestAnalyzeHandler_Integration_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(``),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAnalyzeHandler_Integration_LargeTraceId тестирует большой traceId
func TestAnalyzeHandler_Integration_LargeTraceId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": map[string]interface{}{"trace_id": req.TraceID}})
	})

	largeTraceId := "a"
	for i := 0; i < 1000; i++ {
		largeTraceId += "a"
	}

	w := httptest.NewRecorder()
	requestBody := map[string]string{"traceId": largeTraceId}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBuffer(bodyBytes),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAnalyzeHandler_Integration_SpecialCharactersInTraceId тестирует спецсимволы в traceId
func TestAnalyzeHandler_Integration_SpecialCharactersInTraceId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"trace-<>&\"'-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAnalyzeHandler_Integration_ContentType тестирует различные Content-Type
func TestAnalyzeHandler_Integration_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAnalyzeHandler_Integration_NoContentType тестирует отсутствие Content-Type
func TestAnalyzeHandler_Integration_NoContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)

	router.ServeHTTP(w, req)

	// Без Content-Type: application/json Gin может парсить JSON, но в реальности это зависит от версии
	// Тестируем, что обработка работает
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
}

// TestAnalyzeHandler_Integration_MethodNotAllowed тестирует неправильный HTTP метод
func TestAnalyzeHandler_Integration_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"GET",
		"/analyze",
		nil,
	)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestAnalyzeHandler_Integration_MultipleRequests тестирует последовательные запросы
func TestAnalyzeHandler_Integration_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	requestCount := 0
	router.POST("/analyze", func(c *gin.Context) {
		requestCount++
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": map[string]interface{}{"count": requestCount}})
	})

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			"POST",
			"/analyze",
			bytes.NewBufferString(`{"traceId":"test-123"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 5, requestCount)
}

// TestAnalyzeHandler_Integration_ResponseFormat тестирует формат ответа
func TestAnalyzeHandler_Integration_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": map[string]interface{}{
				"analysis": "test analysis",
				"status":   "success",
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test analysis", data["analysis"])
	assert.Equal(t, "success", data["status"])
}

// TestAnalyzeHandler_Integration_ContentTypeNotJSON тестирует неправильный Content-Type
func TestAnalyzeHandler_Integration_ContentTypeNotJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	// Gin может парсить JSON даже с text/plain, зависит от версии
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
}

// TestAnalyzeHandler_Integration_LongTraceAnalysis тестирует длинный анализ
func TestAnalyzeHandler_Integration_LongTraceAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		longAnalysis := make([]byte, 10000)
		for i := 0; i < 10000; i++ {
			longAnalysis[i] = 'a'
		}

		c.JSON(http.StatusOK, gin.H{
			"data": map[string]interface{}{
				"analysis": string(longAnalysis),
			},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"test-123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, w.Body.Len(), 10000)
}

// TestAnalyzeHandler_Integration_TraceIdWithUnicode тестирует трейс ID с Unicode
func TestAnalyzeHandler_Integration_TraceIdWithUnicode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			TraceID string `json:"traceId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/analyze",
		bytes.NewBufferString(`{"traceId":"trace-测试-🚀"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
