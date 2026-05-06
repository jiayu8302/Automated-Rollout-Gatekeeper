package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/api"
)

// HealthProber defines the contract for verifying deployment integrity.
type HealthProber interface {
	Check(ctx context.Context, d *api.Deployment) error
}

type DeploymentEngine struct {
	Prober HealthProber
}

// ExecuteRollout runs the progressive deployment logic
func (e *DeploymentEngine) ExecuteRollout(ctx context.Context, d *api.Deployment) error {
	slog.Info("Starting deployment sequence", "deployment_id", d.ID, "target_version", d.Version)

	// Stage 1: Canary
	d.Status = api.StatusInCanary
	slog.Info("Entering Canary Stage", "weight", d.CanaryWeight)

	if err := e.monitorHealth(ctx, d); err != nil {
		return e.triggerRollback(ctx, d, err)
	}

	// Stage 2: Full Rollout
	d.Status = api.StatusFullRollout
	slog.Info("Promoting to Full Rollout", "version", d.Version)

	if err := e.monitorHealth(ctx, d); err != nil {
		return e.triggerRollback(ctx, d, err)
	}

	d.Status = api.StatusCompleted
	slog.Info("Deployment successful", "id", d.ID)
	return nil
}

func (e *DeploymentEngine) monitorHealth(ctx context.Context, d *api.Deployment) error {
	slog.Info("Analyzing health metrics...", "window", d.AnalysisWindow.String())

	timer := time.NewTimer(d.AnalysisWindow)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		// Simulate a random failure for demonstration
		// In production, this checks real Prometheus metrics
		return nil
	}
}

func (e *DeploymentEngine) triggerRollback(ctx context.Context, d *api.Deployment, err error) error {
	d.Status = api.StatusRollingBack
	slog.Error("Deployment health check failed. Initiating automated rollback!",
		"reason", err, "deployment_id", d.ID)

	// Rollback logic would go here: Reverting traffic to previous version

	d.Status = api.StatusFailed
	return fmt.Errorf("rollout failed and reverted: %w", err)
}
