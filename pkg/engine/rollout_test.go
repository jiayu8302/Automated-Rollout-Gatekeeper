package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/api"
)

// MockProber simulates various infrastructure health scenarios.
type MockProber struct {
	ShouldFail bool
}

func (m *MockProber) Check(ctx context.Context, d *api.Deployment) error {
	if m.ShouldFail {
		return errors.New("telemetry breach: high error rate")
	}
	return nil
}

func TestExecuteRollout(t *testing.T) {
	config := &api.Deployment{
		ID:             "test-svc-01",
		Version:        "v2.0.0",
		CanaryWeight:   10,
		AnalysisWindow: 5 * time.Millisecond,
	}

	t.Run("Standard Success Path", func(t *testing.T) {
		engine := &DeploymentEngine{
			Prober: &MockProber{ShouldFail: false},
		}

		err := engine.ExecuteRollout(context.Background(), config)
		if err != nil {
			t.Fatalf("Expected success, got: %v", err)
		}
		if config.Status != api.StatusCompleted {
			t.Errorf("Expected StatusCompleted, got %s", config.Status)
		}
	})

	t.Run("Automated Rollback on Failure", func(t *testing.T) {
		failConfig := *config
		engine := &DeploymentEngine{
			Prober: &MockProber{ShouldFail: true},
		}

		err := engine.ExecuteRollout(context.Background(), &failConfig)
		if err == nil {
			t.Fatal("Expected error during rollout, but system proceeded")
		}
		if failConfig.Status != api.StatusFailed {
			t.Errorf("Expected StatusFailed, got %s", failConfig.Status)
		}
	})
}
