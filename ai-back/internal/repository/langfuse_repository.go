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

// LangfuseRepository defines the interface for interacting with Langfuse API.
// It provides methods to fetch trace data from Langfuse with automatic retry logic.
type LangfuseRepository interface {
	// GetTrace fetches trace data from Langfuse by trace ID.
	// It retries up to 3 times with exponential backoff on failure.
	//
	// Параметры:
	//   - ctx: Context for cancellation and timeout
	//   - traceID: Unique identifier of the trace
	//
	// Возвращает:
	//   - Trace data as a map
	//   - Error if all retry attempts fail or trace not found
	GetTrace(ctx context.Context, traceID string) (map[string]interface{}, error)
}

type langfuseClient struct {
	publicKey string
	secretKey string
	baseURL   string
	client    *http.Client
}

// NewLangfuseRepository creates a new Langfuse repository client.
// It initializes HTTP client with 30-second timeout and returns a configured LangfuseRepository instance.
//
// Параметры:
//   - publicKey: Langfuse public API key for authentication
//   - secretKey: Langfuse secret API key for authentication
//   - baseURL: Langfuse API base URL (e.g., https://cloud.langfuse.com)
//
// Возвращает: Configured LangfuseRepository instance.
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
			msg := fmt.Sprintf("failed to create request for trace %s (attempt %d/%d): %w", traceID, attempt, 3, err)
			log.Printf("   ❌ %s", msg)
			lastErr = fmt.Errorf(msg)
			continue
		}
		req.SetBasicAuth(c.publicKey, c.secretKey)

		resp, err := c.client.Do(req)
		if err != nil {
			msg := fmt.Sprintf("failed to fetch trace %s from Langfuse (attempt %d/%d): %w", traceID, attempt, 3, err)
			log.Printf("   ⚠️  %s", msg)
			lastErr = fmt.Errorf(msg)
			continue
		}

		// Обворачиваем в анонимную функцию для корректного закрытия Body
		result, err := func() (map[string]interface{}, error) {
			defer resp.Body.Close()

			log.Printf("   ✅ Статус ответа Langfuse: %d %s", resp.StatusCode, resp.Status)

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				msg := fmt.Sprintf("Langfuse API returned %s for trace %s: %s", resp.Status, traceID, string(bodyBytes))
				log.Printf("   ⚠️  %s", msg)
				return nil, fmt.Errorf(msg)
			}

			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				msg := fmt.Sprintf("failed to decode Langfuse response for trace %s: %w", traceID, err)
				log.Printf("   ❌ %s", msg)
				return nil, fmt.Errorf(msg)
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

	msg := fmt.Sprintf("failed to retrieve trace %s after 3 attempts: %v", traceID, lastErr)
	log.Printf("   ❌ %s", msg)
	return nil, fmt.Errorf(msg)
}
