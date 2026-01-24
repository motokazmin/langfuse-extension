package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"langfuse-analyzer-backend/ai"
	"langfuse-analyzer-backend/internal/service"
)

// MockAIClient - мок для AIClient
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) AnalyzeTrace(ctx context.Context, traceData map[string]interface{}) (string, error) {
	args := m.Called(ctx, traceData)
	return args.String(0), args.Error(1)
}

// MockLangfuseRepository - мок для LangfuseRepository
type MockLangfuseRepository struct {
	mock.Mock
}

func (m *MockLangfuseRepository) GetTrace(ctx context.Context, traceID string) (map[string]interface{}, error) {
	args := m.Called(ctx, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestAnalyzeService_AnalyzeTrace_Success тестирует успешный сценарий с JSON ответом
func TestAnalyzeService_AnalyzeTrace_Success(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{
		"id":   "test-123",
		"name": "test trace",
	}
	mockRepo.On("GetTrace", mock.Anything, "test-123").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return(`{"status": "ok", "result": "analysis"}`, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "test-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Проверяем, что результат распарсен как JSON
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ok", resultMap["status"])
	assert.Equal(t, "analysis", resultMap["result"])

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_SuccessPlainText тестирует успешный сценарий с текстовым ответом
func TestAnalyzeService_AnalyzeTrace_SuccessPlainText(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "test-456"}
	mockRepo.On("GetTrace", mock.Anything, "test-456").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return("Plain text analysis result", nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "test-456")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Проверяем, что результат вернулся как строка
	resultStr, ok := result.(string)
	assert.True(t, ok)
	assert.Equal(t, "Plain text analysis result", resultStr)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_LangfuseError тестирует ошибку получения трейса
func TestAnalyzeService_AnalyzeTrace_LangfuseError(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	mockRepo.On("GetTrace", mock.Anything, "test-123").Return(nil, errors.New("langfuse connection error"))

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "test-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get trace from Langfuse")
	assert.Contains(t, err.Error(), "langfuse connection error")

	mockRepo.AssertExpectations(t)
	// AI не должен быть вызван, если Langfuse вернул ошибку
	mockAI.AssertNotCalled(t, "AnalyzeTrace")
}

// TestAnalyzeService_AnalyzeTrace_AIError тестирует ошибку AI анализа
func TestAnalyzeService_AnalyzeTrace_AIError(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "test-123"}
	mockRepo.On("GetTrace", mock.Anything, "test-123").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return("", errors.New("AI service unavailable"))

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "test-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to analyze trace with AI")
	assert.Contains(t, err.Error(), "AI service unavailable")

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_AIErrorWithCustomError тестирует кастомную AIError
func TestAnalyzeService_AnalyzeTrace_AIErrorWithCustomError(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "test-rate-limit"}
	mockRepo.On("GetTrace", mock.Anything, "test-rate-limit").Return(traceData, nil)

	aiError := &ai.AIError{
		Message:    "Rate limit exceeded",
		StatusCode: 429,
		RetryAfter: 60,
	}
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return("", aiError)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "test-rate-limit")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to analyze trace with AI")

	// Проверяем, что оригинальная ошибка сохранена
	var aiErr *ai.AIError
	assert.ErrorAs(t, err, &aiErr)
	assert.Equal(t, 429, aiErr.StatusCode)
	assert.Equal(t, 60, aiErr.RetryAfter)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_EmptyTraceData тестирует пустые данные трейса
func TestAnalyzeService_AnalyzeTrace_EmptyTraceData(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	emptyTraceData := map[string]interface{}{}
	mockRepo.On("GetTrace", mock.Anything, "empty-trace").Return(emptyTraceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, emptyTraceData).Return(`{"status": "empty"}`, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "empty-trace")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_ComplexTraceData тестирует сложные данные трейса
func TestAnalyzeService_AnalyzeTrace_ComplexTraceData(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	complexTraceData := map[string]interface{}{
		"id":   "complex-trace",
		"name": "Complex Trace",
		"spans": []map[string]interface{}{
			{
				"id":    "span-1",
				"name":  "First Span",
				"input": map[string]interface{}{"key": "value"},
			},
			{
				"id":    "span-2",
				"name":  "Second Span",
				"input": map[string]interface{}{"key2": "value2"},
			},
		},
		"metadata": map[string]interface{}{
			"user":    "test-user",
			"version": "1.0",
		},
	}

	mockRepo.On("GetTrace", mock.Anything, "complex-trace").Return(complexTraceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, complexTraceData).Return(`{"analysis": "complex", "spans_count": 2}`, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "complex-trace")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "complex", resultMap["analysis"])
	assert.Equal(t, float64(2), resultMap["spans_count"])

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_ContextCancellation тестирует отмену контекста
func TestAnalyzeService_AnalyzeTrace_ContextCancellation(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст сразу

	mockRepo.On("GetTrace", mock.Anything, "test-cancel").Return(nil, context.Canceled)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(ctx, "test-cancel")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get trace from Langfuse")

	mockRepo.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_LangfuseNotFound тестирует 404 от Langfuse
func TestAnalyzeService_AnalyzeTrace_LangfuseNotFound(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	mockRepo.On("GetTrace", mock.Anything, "nonexistent").Return(nil, errors.New("Langfuse API returned 404 Not Found for trace nonexistent"))

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "nonexistent")

	mockRepo.AssertExpectations(t)
	mockAI.AssertNotCalled(t, "AnalyzeTrace")
}

// TestAnalyzeService_AnalyzeTrace_LargeResponse тестирует большой ответ от AI
func TestAnalyzeService_AnalyzeTrace_LargeResponse(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "large-response"}
	mockRepo.On("GetTrace", mock.Anything, "large-response").Return(traceData, nil)

	// Создаём большой JSON ответ
	largeResponse := `{"analysis": "` + string(make([]byte, 10000)) + `"}`
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return(largeResponse, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "large-response")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_InvalidJSONFromAI тестирует невалидный JSON от AI
func TestAnalyzeService_AnalyzeTrace_InvalidJSONFromAI(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "invalid-json"}
	mockRepo.On("GetTrace", mock.Anything, "invalid-json").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return("This is not JSON {]", nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "invalid-json")

	// Должен вернуть строку, а не ошибку
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultStr, ok := result.(string)
	assert.True(t, ok)
	assert.Equal(t, "This is not JSON {]", resultStr)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_EmptyAIResponse тестирует пустой ответ от AI
func TestAnalyzeService_AnalyzeTrace_EmptyAIResponse(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "empty-ai-response"}
	mockRepo.On("GetTrace", mock.Anything, "empty-ai-response").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return("", nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "empty-ai-response")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_NestedJSONResponse тестирует вложенный JSON от AI
func TestAnalyzeService_AnalyzeTrace_NestedJSONResponse(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "nested-json"}
	mockRepo.On("GetTrace", mock.Anything, "nested-json").Return(traceData, nil)

	nestedJSON := `{
		"analysis": {
			"summary": "test",
			"details": {
				"errors": ["error1", "error2"],
				"warnings": ["warning1"]
			}
		},
		"metrics": {
			"duration": 1234,
			"tokens": 567
		}
	}`
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return(nestedJSON, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "nested-json")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Contains(t, resultMap, "analysis")
	assert.Contains(t, resultMap, "metrics")

	analysis := resultMap["analysis"].(map[string]interface{})
	assert.Equal(t, "test", analysis["summary"])

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_MultipleTraces тестирует последовательный анализ нескольких трейсов
func TestAnalyzeService_AnalyzeTrace_MultipleTraces(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	// Настраиваем моки для первого трейса
	traceData1 := map[string]interface{}{"id": "trace-1"}
	mockRepo.On("GetTrace", mock.Anything, "trace-1").Return(traceData1, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData1).Return(`{"result": "analysis-1"}`, nil)

	// Настраиваем моки для второго трейса
	traceData2 := map[string]interface{}{"id": "trace-2"}
	mockRepo.On("GetTrace", mock.Anything, "trace-2").Return(traceData2, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData2).Return(`{"result": "analysis-2"}`, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)

	// Анализируем первый трейс
	result1, err1 := svc.AnalyzeTrace(context.Background(), "trace-1")
	assert.NoError(t, err1)
	assert.NotNil(t, result1)
	resultMap1 := result1.(map[string]interface{})
	assert.Equal(t, "analysis-1", resultMap1["result"])

	// Анализируем второй трейс
	result2, err2 := svc.AnalyzeTrace(context.Background(), "trace-2")
	assert.NoError(t, err2)
	assert.NotNil(t, result2)
	resultMap2 := result2.(map[string]interface{})
	assert.Equal(t, "analysis-2", resultMap2["result"])

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_NewAnalyzeService тестирует создание сервиса
func TestAnalyzeService_NewAnalyzeService(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	svc := service.NewAnalyzeService(mockAI, mockRepo)

	assert.NotNil(t, svc)
}

// TestAnalyzeService_AnalyzeTrace_LangfuseTimeout тестирует таймаут Langfuse
func TestAnalyzeService_AnalyzeTrace_LangfuseTimeout(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	mockRepo.On("GetTrace", mock.Anything, "timeout-trace").Return(nil, errors.New("failed to retrieve trace timeout-trace after 3 attempts"))

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "timeout-trace")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "after 3 attempts")

	mockRepo.AssertExpectations(t)
	mockAI.AssertNotCalled(t, "AnalyzeTrace")
}

// TestAnalyzeService_AnalyzeTrace_AIInsufficientCredits тестирует ошибку 402 от AI
func TestAnalyzeService_AnalyzeTrace_AIInsufficientCredits(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "insufficient-credits"}
	mockRepo.On("GetTrace", mock.Anything, "insufficient-credits").Return(traceData, nil)

	aiError := &ai.AIError{
		Message:    "Insufficient credits",
		StatusCode: 402,
	}
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return("", aiError)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "insufficient-credits")

	assert.Error(t, err)
	assert.Nil(t, result)

	var aiErr *ai.AIError
	assert.ErrorAs(t, err, &aiErr)
	assert.Equal(t, 402, aiErr.StatusCode)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_JSONArray тестирует JSON массив в ответе AI
func TestAnalyzeService_AnalyzeTrace_JSONArray(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{"id": "json-array"}
	mockRepo.On("GetTrace", mock.Anything, "json-array").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return(`[{"item": 1}, {"item": 2}]`, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "json-array")

	// JSON массив не парсится как map, должен вернуться как строка
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultStr, ok := result.(string)
	assert.True(t, ok)
	assert.Contains(t, resultStr, "item")

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}

// TestAnalyzeService_AnalyzeTrace_SpecialCharacters тестирует спецсимволы в данных
func TestAnalyzeService_AnalyzeTrace_SpecialCharacters(t *testing.T) {
	mockAI := new(MockAIClient)
	mockRepo := new(MockLangfuseRepository)

	traceData := map[string]interface{}{
		"id":      "special-chars",
		"content": "Test with special chars: <>&\"'",
	}
	mockRepo.On("GetTrace", mock.Anything, "special-chars").Return(traceData, nil)
	mockAI.On("AnalyzeTrace", mock.Anything, traceData).Return(`{"result": "ok"}`, nil)

	svc := service.NewAnalyzeService(mockAI, mockRepo)
	result, err := svc.AnalyzeTrace(context.Background(), "special-chars")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
	mockAI.AssertExpectations(t)
}
