package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/api"
)

// HealthProber defines the contract for analyzing deployment telemetry.
type HealthProber interface {
	Check(ctx context.Context, d *api.Deployment) error
}

// Deployer defines the contract for cloud-specific infrastructure updates.
type Deployer interface {
	Deploy(ctx context.Context, version string, weight int) error
}

// DeploymentEngine orchestrates the progressive delivery state machine.
type DeploymentEngine struct {
	Prober   HealthProber
	Deployer Deployer
	mu       sync.Mutex // Concurrency guard to prevent overlapping rollouts
}

// ExecuteRollout drives the deployment from PENDING to COMPLETED (or FAILED).
func (e *DeploymentEngine) ExecuteRollout(ctx context.Context, d *api.Deployment) error {
	// 1. Concurrency Protection
	if !e.mu.TryLock() {
		return fmt.Errorf("deployment %s rejected: engine is currently busy with another rollout", d.ID)
	}
	defer e.mu.Unlock()

	slog.Info("Starting orchestrated rollout sequence", "id", d.ID, "version", d.Version)

	// --- STAGE 1: Canary Deployment ---
	d.Status = api.StatusInCanary
	slog.Info("Phase 1: Deploying Canary", "weight", d.CanaryWeight)

	if err := e.Deployer.Deploy(ctx, d.Version, d.CanaryWeight); err != nil {
		return e.triggerRollback(ctx, d, fmt.Errorf("canary infrastructure update failed: %w", err))
	}

	if err := e.performAnalysis(ctx, d); err != nil {
		return e.triggerRollback(ctx, d, err)
	}

	// --- STAGE 2: Full Promotion ---
	d.Status = api.StatusFullRollout
	slog.Info("Phase 2: Health gates passed. Promoting to 100% traffic")

	if err := e.Deployer.Deploy(ctx, d.Version, 100); err != nil {
		return e.triggerRollback(ctx, d, fmt.Errorf("full promotion update failed: %w", err))
	}

	if err := e.performAnalysis(ctx, d); err != nil {
		return e.triggerRollback(ctx, d, err)
	}

	// --- STAGE 3: Finalization ---
	d.Status = api.StatusCompleted
	slog.Info("✅ Deployment cycle finished successfully", "id", d.ID)
	return nil
}

// performAnalysis blocks for the analysis window and then runs a health probe.
func (e *DeploymentEngine) performAnalysis(ctx context.Context, d *api.Deployment) error {
	slog.Info("Observing telemetry...", "window", d.AnalysisWindow.String())

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d.AnalysisWindow):
		// Post-observation health check
		if err := e.Prober.Check(ctx, d); err != nil {
			return fmt.Errorf("telemetry analysis failure: %w", err)
		}
	}
	return nil
}

// triggerRollback reverts the system to the Last Known Good (LKG) state.
func (e *DeploymentEngine) triggerRollback(ctx context.Context, d *api.Deployment, cause error) error {
	d.Status = api.StatusRollingBack
	slog.Error("🚨 SAFETY GATE BREACHED: Rolling back immediately",
		"reason", cause.Error(),
		"deployment_id", d.ID)

	// In production, you would fetch the previous version dynamically.
	rollbackErr := e.Deployer.Deploy(ctx, "lkg-version", 100)

	d.Status = api.StatusFailed

	if rollbackErr != nil {
		return fmt.Errorf("FATAL: Rollout failed AND rollback failed: %w (Original Error: %v)", rollbackErr, cause)
	}

	return fmt.Errorf("rollout aborted and successfully reverted: %w", cause)
}
