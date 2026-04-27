package config

import (
	"time"
)

// MetricType defines the category of the data point being evaluated.
type MetricType string

const (
	ErrorRate MetricType = "error_rate"
	Latency   MetricType = "latency_p99"
	CPUUsage  MetricType = "cpu_utilization"
)

// GatingPolicy encapsulates the safety requirements for a deployment to proceed.
type GatingPolicy struct {
	Name            string        `json:"name"`
	EvaluationWindow time.Duration `json:"evaluation_window"` // Time to monitor before making a decision
	Metrics         []Threshold   `json:"metrics"`
}

// Threshold defines the specific limit for a metric.
type Threshold struct {
	Type      MetricType `json:"type"`
	Limit     float64    `json:"limit"`
	IsMinimum bool       `json:"is_minimum"` // true if the value must be ABOVE the limit (e.g., uptime)
}

// GetDefaultProductionPolicy returns a standard set of safety gates for production environments.
func GetDefaultProductionPolicy() GatingPolicy {
	return GatingPolicy{
		Name:             "production-standard-safety-gate",
		EvaluationWindow: 15 * time.Minute,
		Metrics: []Threshold{
			{
				Type:      ErrorRate,
				Limit:     0.001, // 0.1% max error rate
				IsMinimum: false,
			},
			{
				Type:      Latency,
				Limit:     250.0, // 250ms p99 limit
				IsMinimum: false,
			},
		},
	}
}
