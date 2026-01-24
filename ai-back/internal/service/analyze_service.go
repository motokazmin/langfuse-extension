package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"langfuse-analyzer-backend/ai"
	"langfuse-analyzer-backend/internal/repository"
)

// AnalyzeService provides business logic for analyzing traces using AI.
// It orchestrates fetching trace data from Langfuse and analyzing it with AI providers (OpenRouter or Ollama).
type AnalyzeService struct {
	aiClient     ai.AIClient
	langfuseRepo repository.LangfuseRepository
}

// NewAnalyzeService creates a new AnalyzeService instance.
// Параметры:
//   - aiClient: AI client for trace analysis (OpenRouter or Ollama)
//   - langfuseRepo: Repository for fetching traces from Langfuse
//
// Возвращает: Configured AnalyzeService instance.
func NewAnalyzeService(aiClient ai.AIClient, langfuseRepo repository.LangfuseRepository) *AnalyzeService {
	return &AnalyzeService{
		aiClient:     aiClient,
		langfuseRepo: langfuseRepo,
	}
}

// AnalyzeTrace analyzes a trace by its ID using AI.
// It first fetches the trace from Langfuse, then sends it to AI for analysis.
// The AI response is parsed as JSON if possible, otherwise returned as plain text.
//
// Параметры:
//   - ctx: Context for cancellation and timeout
//   - traceID: Unique identifier of the trace to analyze
//
// Возвращает:
//   - Analysis result from AI (map[string]interface{} for JSON responses, string for plain text)
//   - Error if trace fetch or AI analysis fails
func (s *AnalyzeService) AnalyzeTrace(ctx context.Context, traceID string) (interface{}, error) {
	log.Println("----------------------------------------------")
	log.Println("🔄 ШАГ 1: Получение данных трейса из Langfuse")

	// 1. Получаем трейс из Langfuse
	traceData, err := s.langfuseRepo.GetTrace(ctx, traceID)
	if err != nil {
		log.Printf("❌ Ошибка получения трейса: %v", err)
		return nil, fmt.Errorf("failed to get trace from Langfuse: %w", err)
	}

	log.Printf("✅ Трейс получен, размер данных: %d байт", len(fmt.Sprintf("%v", traceData)))
	log.Println("----------------------------------------------")
	log.Println("🤖 ШАГ 2: Отправка на анализ AI")

	// 2. Анализируем через AI
	analysisResult, err := s.aiClient.AnalyzeTrace(ctx, traceData)
	if err != nil {
		log.Printf("❌ Ошибка анализа AI: %v", err)
		return nil, fmt.Errorf("failed to analyze trace with AI: %w", err)
	}

	log.Printf("✅ AI анализ завершён, длина ответа: %d символов", len(analysisResult))

	// 3. Парсим результат
	var structuredResponse map[string]interface{}
	if err := json.Unmarshal([]byte(analysisResult), &structuredResponse); err != nil {
		log.Println("⚠️  Ответ не в формате JSON, отправляем как строку")
		return analysisResult, nil
	}

	log.Println("✅ Ответ распарсен как JSON")
	return structuredResponse, nil
}
