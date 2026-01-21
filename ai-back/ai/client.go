package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AIError - специальная ошибка с дополнительной информацией
type AIError struct {
	StatusCode int
	Message    string
	RetryAfter int // секунды
}

func (e *AIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry after %d seconds)", e.Message, e.RetryAfter)
	}
	return e.Message
}

// AIClient - интерфейс для работы с различными AI провайдерами
type AIClient interface {
	AnalyzeTrace(ctx context.Context, traceData map[string]interface{}) (string, error)
}

// ProviderType - тип провайдера AI
type ProviderType string

const (
	ProviderOpenRouter ProviderType = "openrouter"
	ProviderOllama     ProviderType = "ollama"
)

// OpenAIClient - клиент для работы с OpenAI-совместимыми API (OpenRouter)
type OpenAIClient struct {
	client    *openai.Client
	model     string
	maxTokens int
}

// OllamaClient - клиент для работы с Ollama
type OllamaClient struct {
	baseURL   string
	model     string
	maxTokens int
	client    *http.Client
}

// headerTransport добавляет кастомные заголовки для OpenRouter
type headerTransport struct {
	base http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("HTTP-Referer", "http://localhost")
	return t.base.RoundTrip(req)
}

// NewAIClient создает подходящего клиента на основе конфигурации
func NewAIClient(provider ProviderType, apiKey, baseURL, model string, maxTokens int) AIClient {
	switch provider {
	case ProviderOllama:
		return NewOllamaClient(baseURL, model, maxTokens)
	default:
		return NewOpenAIClient(apiKey, baseURL, model, maxTokens)
	}
}

// NewOpenAIClient создает нового клиента для OpenRouter
func NewOpenAIClient(apiKey, baseURL, model string, maxTokens int) *OpenAIClient {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	// Создаем кастомный HTTP-клиент с нужными заголовками для OpenRouter
	transport := &headerTransport{
		base: http.DefaultTransport,
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	config.HTTPClient = httpClient

	client := openai.NewClientWithConfig(config)

	// Устанавливаем модель по умолчанию, если она не указана
	if model == "" {
		model = "google/gemini-2.0-flash-exp:free"
	}

	// Устанавливаем maxTokens по умолчанию, если не указан
	if maxTokens <= 0 {
		maxTokens = 1000
	}

	return &OpenAIClient{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}
}

// NewOllamaClient создает нового клиента для Ollama
func NewOllamaClient(baseURL, model string, maxTokens int) *OllamaClient {
	// Устанавливаем baseURL по умолчанию для Ollama
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// Устанавливаем модель по умолчанию
	if model == "" {
		model = "llama3.2"
	}

	// Устанавливаем maxTokens по умолчанию
	if maxTokens <= 0 {
		maxTokens = 1000
	}

	// Читаем таймаут из переменной окружения (если не указан - 120 секунд)
	timeout := 120 * time.Second
	if timeoutStr := os.Getenv("OLLAMA_TIMEOUT"); timeoutStr != "" {
		if timeoutSec, err := strconv.Atoi(timeoutStr); err == nil && timeoutSec > 0 {
			timeout = time.Duration(timeoutSec) * time.Second
		}
	}

	return &OllamaClient{
		baseURL:   baseURL,
		model:     model,
		maxTokens: maxTokens,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// AnalyzeTrace - анализ трейса через OpenRouter
func (c *OpenAIClient) AnalyzeTrace(ctx context.Context, traceData map[string]interface{}) (string, error) {
	traceStr, err := json.Marshal(traceData)
	if err != nil {
		return "", fmt.Errorf("ошибка при маршалинге traceData: %w", err)
	}

	systemPrompt := getSystemPrompt()
	userPrompt := fmt.Sprintf("Проанализируй следующий JSON-трейс: %s", traceStr)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			MaxTokens: c.maxTokens,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		var apiErr *openai.APIError
		if errors.As(err, &apiErr) {
			aiErr := &AIError{
				StatusCode: apiErr.HTTPStatusCode,
				Message:    fmt.Sprintf("ошибка при вызове ChatCompletion: status %d, message: %s", apiErr.HTTPStatusCode, apiErr.Message),
				RetryAfter: 0,
			}

			if apiErr.HTTPStatusCode == 429 {
				if retrySeconds := extractRetryAfter(apiErr.Message); retrySeconds > 0 {
					aiErr.RetryAfter = retrySeconds
				} else {
					aiErr.RetryAfter = 10
				}
			}

			return "", aiErr
		}

		return "", fmt.Errorf("ошибка при вызове ChatCompletion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("нет ответа от AI")
	}

	return resp.Choices[0].Message.Content, nil
}

// OllamaRequest - структура запроса к Ollama API
type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Format   string          `json:"format,omitempty"`
	Options  *OllamaOptions  `json:"options,omitempty"`
}

// OllamaMessage - сообщение в Ollama
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaOptions - опции для Ollama
type OllamaOptions struct {
	NumPredict int `json:"num_predict,omitempty"` // аналог max_tokens
}

// OllamaResponse - ответ от Ollama API
type OllamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

// AnalyzeTrace - анализ трейса через Ollama
func (c *OllamaClient) AnalyzeTrace(ctx context.Context, traceData map[string]interface{}) (string, error) {
	traceStr, err := json.Marshal(traceData)
	if err != nil {
		return "", fmt.Errorf("ошибка при маршалинге traceData: %w", err)
	}

	systemPrompt := getSystemPrompt()
	userPrompt := fmt.Sprintf("Проанализируй следующий JSON-трейс: %s", traceStr)

	// Формируем запрос к Ollama
	reqBody := OllamaRequest{
		Model: c.model,
		Messages: []OllamaMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Stream: false,
		Format: "json", // Просим Ollama возвращать JSON
		Options: &OllamaOptions{
			NumPredict: c.maxTokens,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка при маршалинге запроса к Ollama: %w", err)
	}

	// Отправляем запрос к Ollama
	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса к Ollama: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", &AIError{
			StatusCode: http.StatusServiceUnavailable,
			Message:    fmt.Sprintf("ошибка при подключении к Ollama: %v. Убедитесь, что Ollama запущена на %s", err, c.baseURL),
			RetryAfter: 0,
		}
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", &AIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Ollama вернула ошибку %d: %s", resp.StatusCode, string(bodyBytes)),
			RetryAfter: 0,
		}
	}

	// Читаем ответ
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа от Ollama: %w", err)
	}

	if !ollamaResp.Done {
		return "", fmt.Errorf("Ollama вернула неполный ответ")
	}

	return ollamaResp.Message.Content, nil
}

// getSystemPrompt возвращает системный промпт для анализа
func getSystemPrompt() string {
	return `
Ты — 'TraceDebugger', элитный AI-аналитик, специализирующийся на поиске проблем в логах выполнения LLM-приложений. 

**ВАЖНО: Отвечай ТОЛЬКО на русском языке!**

Твоя задача — проанализировать предоставленный JSON-трейс из системы Langfuse и дать четкий, структурированный отчет **НА РУССКОМ ЯЗЫКЕ**.

# Инструкции:
1.  **Изучи общую информацию:** Обрати внимание на общую задержку ('latency') и стоимость ('totalCost') всего трейса.
2.  **Проанализируй шаги ('observations'):** Внимательно изучи каждый шаг в массиве 'observations'.
3.  **Выяви аномалии:** Найди одну из следующих проблем: 'ERROR' (ошибка), 'PERFORMANCE_BOTTLENECK' (узкое место производительности), 'HIGH_COST' (высокая стоимость), 'LOGICAL_LOOP' (логический цикл).
4.  **Сформируй отчет НА РУССКОМ ЯЗЫКЕ:** Предоставь свой вывод в строго определенном JSON-формате. Не добавляй никаких комментариев или текста вне этого JSON.

# Формат вывода (обязателен, все тексты на русском):
{
  "analysisSummary": {
    "traceId": "ID_ТРЕЙСА",
    "overallStatus": "HEALTHY | WARNING | ERROR",
    "keyFinding": "Ключевой вывод в одном предложении на русском языке."
  },
  "detailedAnalysis": {
    "anomalyType": "NONE | ERROR | PERFORMANCE_BOTTLENECK | HIGH_COST | LOGICAL_LOOP",
    "description": "Подробное описание найденной проблемы на русском языке.",
    "rootCause": "Твоя гипотеза о первопричине проблемы на русском языке.",
    "recommendation": "Конкретный, действенный совет для разработчика на русском языке."
  }
}

**Все поля description, rootCause, recommendation и keyFinding должны быть заполнены текстом на русском языке!**
`
}

// extractRetryAfter пытается найти retry время в сообщении об ошибке
func extractRetryAfter(message string) int {
	message = strings.ToLower(message)

	// Паттерн 1: "retry after X seconds"
	if idx := strings.Index(message, "retry after"); idx >= 0 {
		substr := message[idx+11:]
		var seconds int
		if _, err := fmt.Sscanf(substr, "%d", &seconds); err == nil {
			return seconds
		}
	}

	// Паттерн 2: "retry in Xs" или "retry in X seconds"
	if idx := strings.Index(message, "retry in"); idx >= 0 {
		substr := message[idx+8:]
		var seconds int
		if _, err := fmt.Sscanf(substr, "%d", &seconds); err == nil {
			return seconds
		}
	}

	// Паттерн 3: число в секундах где-то в сообщении
	words := strings.Fields(message)
	for i, word := range words {
		if seconds, err := strconv.Atoi(word); err == nil && seconds > 0 && seconds < 3600 {
			if i > 0 {
				prev := words[i-1]
				if strings.Contains(prev, "second") || strings.Contains(prev, "sec") ||
					strings.Contains(prev, "retry") || strings.Contains(prev, "wait") {
					return seconds
				}
			}
			if i < len(words)-1 {
				next := words[i+1]
				if strings.Contains(next, "second") || strings.Contains(next, "sec") {
					return seconds
				}
			}
		}
	}

	return 0
}
