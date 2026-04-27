package provider

import (
	"context"
	"time"
)

// MetricData represents a point-in-time telemetry value.
type MetricData struct {
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// MetricProvider defines the interface for external monitoring systems.
// This abstraction allows the Gatekeeper to support Prometheus, Azure Monitor, etc.
type MetricProvider interface {
	// GetCurrentValue retrieves the most recent data point for a specific metric name.
	GetCurrentValue(ctx context.Context, metricName string, query string) (*MetricData, error)

	// GetAverageInRange calculates the mean value over a specified duration.
	GetAverageInRange(ctx context.Context, metricName string, window time.Duration) (float64, error)
}

// MockProvider is used for unit testing and local development without external dependencies.
type MockProvider struct{}

func (m *MockProvider) GetCurrentValue(ctx context.Context, metricName string, query string) (*MetricData, error) {
	return &MetricData{
		Value:     0.0005, // Simulated healthy error rate
		Timestamp: time.Now(),
	}, nil
}

func (m *MockProvider) GetAverageInRange(ctx context.Context, window time.Duration) (float64, error) {
	return 0.0005, nil
}
