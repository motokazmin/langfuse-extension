package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"langfuse-analyzer-backend/ai"
	"langfuse-analyzer-backend/internal/repository"
)

type AnalyzeService struct {
	aiClient     ai.AIClient
	langfuseRepo repository.LangfuseRepository
}

func NewAnalyzeService(aiClient ai.AIClient, langfuseRepo repository.LangfuseRepository) *AnalyzeService {
	return &AnalyzeService{
		aiClient:     aiClient,
		langfuseRepo: langfuseRepo,
	}
}

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
