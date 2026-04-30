package provider

import (
	"context"
	"time"
	"math"
	"log"
)

// FetchWithRetry attempts to get metrics with an exponential backoff strategy.
func FetchWithRetry(ctx context.Context, provider MetricProvider, metric string) (float64, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		data, err := provider.GetCurrentValue(ctx, metric, "")
		if err == nil {
			return data.Value, nil
		}

		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Printf("Metric fetch failed, retrying in %v... (Attempt %d/%d)", backoff, i+1, maxRetries)
		
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
	return 0, fmt.Errorf("failed to fetch metric %s after %d retries", metric, maxRetries)
}
