package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"langfuse-analyzer-backend/ai"
	"langfuse-analyzer-backend/internal/handler"
	"langfuse-analyzer-backend/internal/repository"
	"langfuse-analyzer-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	// ====================================================================
	// DEPENDENCY INJECTION
	// ====================================================================
	// Создаём AI клиента
	aiClient := ai.NewAIClient(provider, apiKey, baseURL, aiModel, maxTokens)
	log.Println("✅ AI клиент успешно инициализирован")

	// Создаём Langfuse репозиторий
	langfuseRepo := repository.NewLangfuseRepository(
		os.Getenv("LANGFUSE_PUBLIC_KEY"),
		os.Getenv("LANGFUSE_SECRET_KEY"),
		os.Getenv("LANGFUSE_BASEURL"),
	)
	log.Println("✅ Langfuse репозиторий инициализирован")

	// Создаём сервис
	analyzeService := service.NewAnalyzeService(aiClient, langfuseRepo)
	log.Println("✅ Analyze сервис инициализирован")

	// Создаём handler
	analyzeHandler := handler.NewAnalyzeHandler(analyzeService)
	log.Println("✅ Analyze handler инициализирован")

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
	router.POST("/analyze", analyzeHandler.Handle)

	log.Println("==============================================")
	log.Println("🚀 Go-сервис запущен на http://localhost:8080")
	log.Println("==============================================")
	router.Run(":8080")
}
