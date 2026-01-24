package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type LangfuseRepository interface {
	GetTrace(ctx context.Context, traceID string) (map[string]interface{}, error)
}

type langfuseClient struct {
	publicKey string
	secretKey string
	baseURL   string
	client    *http.Client
}

func NewLangfuseRepository(publicKey, secretKey, baseURL string) LangfuseRepository {
	return &langfuseClient{
		publicKey: publicKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *langfuseClient) GetTrace(ctx context.Context, traceID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/public/traces/%s", c.baseURL, traceID)
	log.Printf("   🌐 Запрос к Langfuse API: %s", url)

	// Retry логика - 3 попытки
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			log.Printf("   🔄 Попытка %d/3", attempt)
			time.Sleep(time.Duration(attempt) * time.Second) // Экспоненциальная задержка
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			log.Printf("   ❌ Ошибка создания запроса (попытка %d): %v", attempt, err)
			lastErr = err
			continue
		}
		req.SetBasicAuth(c.publicKey, c.secretKey)

		resp, err := c.client.Do(req)
		if err != nil {
			log.Printf("   ⚠️  Ошибка HTTP запроса (попытка %d): %v", attempt, err)
			lastErr = err
			continue
		}

		// Обворачиваем в анонимную функцию для корректного закрытия Body
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
