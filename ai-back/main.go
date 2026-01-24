package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"langfuse-analyzer-backend/ai"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type AnalyzeRequest struct {
	TraceID string `json:"traceId"`
}

var aiClient ai.AIClient

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: не удалось загрузить .env файл.")
	}

	// ====================================================================
	// ОПРЕДЕЛЕНИЕ AI ПРОВАЙДЕРА
	// ====================================================================
	aiProvider := strings.ToLower(os.Getenv("AI_PROVIDER"))
	if aiProvider == "" {
		aiProvider = "openrouter" // По умолчанию OpenRouter
	}

	var provider ai.ProviderType
	switch aiProvider {
	case "ollama":
		provider = ai.ProviderOllama
		log.Println("🤖 Используется AI провайдер: OLLAMA")
	case "openrouter":
		provider = ai.ProviderOpenRouter
		log.Println("🤖 Используется AI провайдер: OPENROUTER")
	default:
		log.Fatalf("❌ Неизвестный AI провайдер: %s. Доступные: openrouter, ollama", aiProvider)
	}

	// ====================================================================
	// КОНФИГУРАЦИЯ AI КЛИЕНТА
	// ====================================================================
	var apiKey, baseURL, aiModel string
	var maxTokens int

	if provider == ai.ProviderOllama {
		// Для Ollama API ключ не нужен
		baseURL = os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434"
			log.Printf("OLLAMA_BASE_URL не указан, используем по умолчанию: %s", baseURL)
		}

		aiModel = os.Getenv("OLLAMA_MODEL")
		if aiModel == "" {
			aiModel = "llama3.2"
			log.Printf("OLLAMA_MODEL не указана, используем по умолчанию: %s", aiModel)
		}

		// Читаем таймаут для Ollama
		ollamaTimeout := os.Getenv("OLLAMA_TIMEOUT")
		if ollamaTimeout == "" {
			ollamaTimeout = "120"
		}
		log.Printf("⏱️  Таймаут Ollama: %s секунд", ollamaTimeout)

		log.Printf("📍 Ollama URL: %s", baseURL)
		log.Printf("🧠 Модель Ollama: %s", aiModel)

	} else {
		// Для OpenRouter нужен API ключ
		apiKey = os.Getenv("AI_API_KEY")
		if apiKey == "" {
			log.Fatal("❌ Переменная окружения AI_API_KEY не установлена (требуется для OpenRouter).")
		}

		baseURL = os.Getenv("AI_BASE_URL")
		if baseURL == "" {
			baseURL = "https://openrouter.ai/api/v1"
		}

		aiModel = os.Getenv("AI_MODEL")
		if aiModel == "" {
			aiModel = "google/gemini-2.0-flash-exp:free"
			log.Printf("AI_MODEL не указана, используем по умолчанию: %s", aiModel)
		}

		log.Printf("🧠 Модель OpenRouter: %s", aiModel)
	}

	// Читаем max_tokens из переменных окружения (по умолчанию 1000)
	maxTokensStr := os.Getenv("AI_MAX_TOKENS")
	maxTokens = 1000
	if maxTokensStr != "" {
		if parsed, err := strconv.Atoi(maxTokensStr); err == nil {
			maxTokens = parsed
		} else {
			log.Printf("⚠️ Неверное значение AI_MAX_TOKENS: %s, используем 1000", maxTokensStr)
		}
	}
	log.Printf("📊 Максимум токенов для AI: %d", maxTokens)

	// Создаём AI клиента
	aiClient = ai.NewAIClient(provider, apiKey, baseURL, aiModel, maxTokens)
	log.Println("✅ AI клиент успешно инициализирован")

	// ====================================================================
	// НАСТРОЙКА CHROME EXTENSION CORS
	// ====================================================================
	chromeExtensionID := os.Getenv("CHROME_EXTENSION_ID")
	if chromeExtensionID == "" {
		log.Fatal("❌ Переменная окружения CHROME_EXTENSION_ID не установлена.")
	}

	router := gin.Default()

	chromeExtensionOrigin := "chrome-extension://" + chromeExtensionID
	log.Printf("🔐 Разрешаем CORS для: %s", chromeExtensionOrigin)

	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			allowed := origin == chromeExtensionOrigin
			if allowed {
				log.Printf("CORS: Разрешен запрос от %s", origin)
			} else {
				log.Printf("CORS: Отклонен запрос от %s", origin)
			}
			return allowed
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(config))

	// ====================================================================
	// РОУТЫ
	// ====================================================================
	router.POST("/analyze", handleAnalyzeRequest)

	log.Println("==============================================")
	log.Println("🚀 Go-сервис запущен на http://localhost:8080")
	log.Println("==============================================")
	router.Run(":8080")
}

func handleAnalyzeRequest(c *gin.Context) {
	log.Println("==============================================")
	log.Println("📥 НОВЫЙ ЗАПРОС НА АНАЛИЗ")
	log.Println("==============================================")
	log.Printf("Origin: %s", c.Request.Header.Get("Origin"))
	log.Printf("Method: %s", c.Request.Method)
	log.Printf("Content-Type: %s", c.Request.Header.Get("Content-Type"))

	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	log.Printf("✅ Получен запрос на анализ traceId: %s", req.TraceID)
	log.Println("----------------------------------------------")
	log.Println("🔄 ШАГ 1: Получение данных трейса из Langfuse")

	traceData, err := getTraceFromLangfuse(req.TraceID)
	if err != nil {
		log.Printf("❌ Ошибка получения трейса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trace from Langfuse: " + err.Error()})
		return
	}

	log.Printf("✅ Трейс получен, размер данных: %d байт", len(fmt.Sprintf("%v", traceData)))
	log.Println("----------------------------------------------")
	log.Println("🤖 ШАГ 2: Отправка на анализ AI")

	analysisResult, err := aiClient.AnalyzeTrace(c.Request.Context(), traceData)
	if err != nil {
		log.Printf("❌ Ошибка анализа AI: %v", err)

		// Проверяем если это наша кастомная AIError
		var aiErr *ai.AIError
		if errors.As(err, &aiErr) {
			switch aiErr.StatusCode {
			case 429:
				log.Printf("⚠️  Rate limit от AI провайдера, retry after %d секунд", aiErr.RetryAfter)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":      "Слишком много запросов к AI. Попробуйте позже.",
					"code":       "RATE_LIMIT",
					"retryAfter": aiErr.RetryAfter,
				})
				return
			case 402:
				log.Println("⚠️  Недостаточно кредитов на AI провайдере")
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": "Недостаточно кредитов для AI анализа. Пополните баланс на OpenRouter.",
					"code":  "INSUFFICIENT_CREDITS",
				})
				return
			case 503:
				log.Println("⚠️  AI сервис недоступен")
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": aiErr.Message,
					"code":  "SERVICE_UNAVAILABLE",
				})
				return
			default:
				c.JSON(aiErr.StatusCode, gin.H{
					"error": aiErr.Message,
				})
				return
			}
		}

		// Проверяем тип ошибки по тексту (fallback для старых ошибок)
		errorMsg := err.Error()

		// 429 Too Many Requests - rate limit
		if contains(errorMsg, "429") || contains(errorMsg, "Too Many Requests") || contains(errorMsg, "rate limit") {
			log.Println("⚠️  Rate limit от AI провайдера, возвращаем 429")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Слишком много запросов к AI. Попробуйте через несколько секунд.",
				"code":       "RATE_LIMIT",
				"retryAfter": 10,
			})
			return
		}

		// 402 Payment Required - недостаточно кредитов
		if contains(errorMsg, "402") || contains(errorMsg, "credits") || contains(errorMsg, "Payment Required") {
			log.Println("⚠️  Недостаточно кредитов на AI провайдере")
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": "Недостаточно кредитов для AI анализа. Пополните баланс на OpenRouter.",
				"code":  "INSUFFICIENT_CREDITS",
			})
			return
		}

		// Остальные ошибки - 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze trace with LLM: " + err.Error()})
		return
	}

	log.Printf("✅ AI анализ завершён, длина ответа: %d символов", len(analysisResult))
	log.Println("----------------------------------------------")
	log.Println("📤 ШАГ 3: Отправка результата в браузер")

	var structuredResponse map[string]interface{}
	if err := json.Unmarshal([]byte(analysisResult), &structuredResponse); err != nil {
		log.Println("⚠️  Ответ не в формате JSON, отправляем как строку")
		c.JSON(http.StatusOK, gin.H{"data": analysisResult})
	} else {
		log.Println("✅ Ответ распарсен как JSON")
		c.JSON(http.StatusOK, gin.H{"data": structuredResponse})
	}

	log.Println("==============================================")
	log.Println("✅ ЗАПРОС УСПЕШНО ОБРАБОТАН")
	log.Println("==============================================")
	log.Println()
}

// contains проверяет содержится ли подстрока в строке (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getTraceFromLangfuse(traceID string) (map[string]interface{}, error) {
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	host := os.Getenv("LANGFUSE_BASEURL")

	url := fmt.Sprintf("%s/api/public/traces/%s", host, traceID)
	log.Printf("   🌐 Запрос к Langfuse API: %s", url)

	// Увеличиваем таймаут до 30 секунд
	client := &http.Client{Timeout: 30 * time.Second}

	// Retry логика - 3 попытки
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			log.Printf("   🔄 Попытка %d/3", attempt)
			time.Sleep(time.Duration(attempt) * time.Second) // Экспоненциальная задержка
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("   ❌ Ошибка создания запроса (попытка %d): %v", attempt, err)
			lastErr = err
			continue
		}
		req.SetBasicAuth(publicKey, secretKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("   ⚠️  Ошибка HTTP запроса (попытка %d): %v", attempt, err)
			lastErr = err
			continue
		}

		// ✅ ИСПРАВЛЕНИЕ: Обворачиваем в анонимную функцию для корректного закрытия Body
		// при каждой итерации, а не в конце всей функции
		result, err := func() (map[string]interface{}, error) {
			defer resp.Body.Close()

			log.Printf("   ✅ Статус ответа Langfuse: %d %s", resp.StatusCode, resp.Status)

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				log.Printf("   ⚠️  Тело ответа: %s", string(bodyBytes))
				return nil, fmt.Errorf("Langfuse API вернул статус: %s", resp.Status)
			}

			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				log.Printf("   ❌ Ошибка декодирования JSON: %v", err)
				return nil, err
			}

			log.Printf("   ✅ Данные трейса успешно получены")
			return data, nil
		}()

		if result != nil {
			return result, nil
		}

		if err != nil {
			lastErr = err
		}
	}

	log.Printf("   ❌ Все попытки исчерпаны")
	return nil, lastErr
}
