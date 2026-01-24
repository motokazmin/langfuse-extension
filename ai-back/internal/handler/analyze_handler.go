package handler

import (
	"errors"
	"log"
	"net/http"

	"langfuse-analyzer-backend/ai"
	"langfuse-analyzer-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyzeHandler struct {
	analyzeService *service.AnalyzeService
}

func NewAnalyzeHandler(analyzeService *service.AnalyzeService) *AnalyzeHandler {
	return &AnalyzeHandler{
		analyzeService: analyzeService,
	}
}

type AnalyzeRequest struct {
	TraceID string `json:"traceId" binding:"required"`
}

func (h *AnalyzeHandler) Handle(c *gin.Context) {
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

	result, err := h.analyzeService.AnalyzeTrace(c.Request.Context(), req.TraceID)
	if err != nil {
		log.Printf("❌ Ошибка анализа: %v", err)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze trace: " + err.Error()})
		return
	}

	log.Println("----------------------------------------------")
	log.Println("📤 ШАГ 3: Отправка результата в браузер")
	c.JSON(http.StatusOK, gin.H{"data": result})

	log.Println("==============================================")
	log.Println("✅ ЗАПРОС УСПЕШНО ОБРАБОТАН")
	log.Println("==============================================")
	log.Println()
}

// contains проверяет содержится ли подстрока в строке
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
