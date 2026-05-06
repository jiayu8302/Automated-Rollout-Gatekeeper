package api

import "time"

type RolloutStatus string

const (
	StatusPending     RolloutStatus = "PENDING"
	StatusInCanary    RolloutStatus = "IN_CANARY"
	StatusFullRollout RolloutStatus = "FULL_ROLLOUT"
	StatusRollingBack RolloutStatus = "ROLLING_BACK"
	StatusCompleted   RolloutStatus = "COMPLETED"
	StatusFailed      RolloutStatus = "FAILED"
)

// Deployment represents a specific version rollout task.
type Deployment struct {
	ID             string        `yaml:"id"`
	Version        string        `yaml:"version"`
	Strategy       string        `yaml:"strategy"` // e.g., "canary", "blue-green"
	Status         RolloutStatus `yaml:"status"`
	ErrorThreshold float64       `yaml:"error_threshold"` // Max allowed error rate (e.g., 0.05)
	CanaryWeight   int           `yaml:"canary_weight"`   // Percentage for initial stage
	AnalysisWindow time.Duration `yaml:"analysis_window"` // Time to wait before next stage
}
