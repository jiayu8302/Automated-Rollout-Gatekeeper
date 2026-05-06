package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/api"
)

// --- Mocks ---

type MockProber struct {
	ShouldFail bool
}

func (m *MockProber) Check(ctx context.Context, d *api.Deployment) error {
	if m.ShouldFail {
		return errors.New("high error rate detected")
	}
	return nil
}

type MockDeployer struct {
	CallCount int
}

func (m *MockDeployer) Deploy(ctx context.Context, v string, w int) error {
	m.CallCount++
	return nil
}

// --- Tests ---

func TestExecuteRollout(t *testing.T) {
	config := &api.Deployment{
		ID:             "unit-test-deploy",
		Version:        "v2.0.1",
		CanaryWeight:   20,
		AnalysisWindow: 5 * time.Millisecond, // Fast tests
	}

	t.Run("Path: Full Success", func(t *testing.T) {
		deployer := &MockDeployer{}
		engine := &DeploymentEngine{
			Prober:   &MockProber{ShouldFail: false},
			Deployer: deployer,
		}

		err := engine.ExecuteRollout(context.Background(), config)

		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}
		if config.Status != api.StatusCompleted {
			t.Errorf("Expected StatusCompleted, got %s", config.Status)
		}
		// Expect 2 calls: one for Canary, one for Full Promotion
		if deployer.CallCount != 2 {
			t.Errorf("Expected 2 deployment calls, got %d", deployer.CallCount)
		}
	})

	t.Run("Path: Canary Failure and Rollback", func(t *testing.T) {
		failConfig := *config // Copy to avoid state leakage
		deployer := &MockDeployer{}
		engine := &DeploymentEngine{
			Prober:   &MockProber{ShouldFail: true},
			Deployer: deployer,
		}

		err := engine.ExecuteRollout(context.Background(), &failConfig)

		if err == nil {
			t.Error("Expected error during rollout, but got nil")
		}
		if failConfig.Status != api.StatusFailed {
			t.Errorf("Expected StatusFailed after rollback, got %s", failConfig.Status)
		}
		// Expect 2 calls: one for Canary, one for the Rollback itself
		if deployer.CallCount != 2 {
			t.Errorf("Expected 2 deployment calls (one being rollback), got %d", deployer.CallCount)
		}
	})

	t.Run("Concurrency Protection", func(t *testing.T) {
		engine := &DeploymentEngine{
			Prober:   &MockProber{},
			Deployer: &MockDeployer{},
		}

		// Lock the engine manually
		engine.mu.Lock()
		defer engine.mu.Unlock()

		err := engine.ExecuteRollout(context.Background(), config)
		if err == nil || err.Error() == "" {
			t.Error("Expected error for concurrent rollout attempt, but got nil")
		}
	})
}
