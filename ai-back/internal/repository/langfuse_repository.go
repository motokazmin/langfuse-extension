package repository

import (
	"context"
)

// LangfuseRepository defines the interface for interacting with Langfuse API.
// It provides methods to fetch trace data from Langfuse.
type LangfuseRepository interface {
	// GetTrace fetches trace data from Langfuse by trace ID.
	// It retries up to 3 times with exponential backoff on failure.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - traceID: Unique identifier of the trace
	//
	// Returns:
	//   - Trace data as a map
	//   - Error if all retry attempts fail or trace not found
	GetTrace(ctx context.Context, traceID string) (map[string]interface{}, error)
}
