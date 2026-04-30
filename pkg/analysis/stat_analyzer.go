package analysis

import (
	"fmt"
	"math"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/config"
)

// AnalysisResult represents the outcome of a reliability evaluation.
type AnalysisResult struct {
	IsSafe     bool
	Confidence float64
	Reason     string
}

// CompareCanaryToBaseline performs a delta analysis between two versions.
func (a *RolloutAnalyzer) CompareCanaryToBaseline(baseline, canary float64) AnalysisResult {
	// Calculate the relative degradation
	// If canary error rate is significantly higher than baseline, we halt.
	if baseline == 0 {
		baseline = 0.0001 // Prevent division by zero
	}

	degradation := (canary - baseline) / baseline

	// If error rate increased by more than 20%, it's an anomaly.
	threshold := 0.20 
	if degradation > threshold {
		return AnalysisResult{
			IsSafe:     false,
			Confidence: 0.95,
			Reason:     fmt.Sprintf("Canary degradation (%.2f%%) exceeds safety threshold (%.2f%%)", degradation*100, threshold*100),
		}
	}

	return AnalysisResult{
		IsSafe:     true,
		Confidence: 1.0,
		Reason:     "Canary performance is within acceptable variance of baseline.",
	}
}
