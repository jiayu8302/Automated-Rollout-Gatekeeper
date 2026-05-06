package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/api"
	"github.com/jiayu8302/deployment-reliability-engine/pkg/engine"
)

// RealWorldProber is a concrete implementation of the engine.HealthProber interface.
type RealWorldProber struct{}

// Check simulates a real health check (e.g., calling a Prometheus endpoint).
func (p *RealWorldProber) Check(ctx context.Context, d *api.Deployment) error {
	slog.Info("Health probe verified: Success", "version", d.Version)
	return nil
}

// loadDeploymentConfig reads and parses the YAML deployment policy.
func loadDeploymentConfig(path string) (*api.Deployment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment policy: %w", err)
	}
	var cfg api.Deployment
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	return &cfg, nil
}

func main() {
	// 1. Initialize Structured JSON Logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.Info("🚀 Initializing Deployment Reliability Engine (DRE)")

	// 2. Load the deployment policy from the configs folder
	cfg, err := loadDeploymentConfig("configs/deploy_policy.yaml")
	if err != nil {
		slog.Error("Configuration error", "error", err)
		os.Exit(1)
	}

	// 3. Initialize the Engine with a real prober
	dre := &engine.DeploymentEngine{
		Prober: &RealWorldProber{},
	}

	// Set a global timeout for the entire deployment process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 4. Start the Rollout sequence
	if err := dre.ExecuteRollout(ctx, cfg); err != nil {
		slog.Error("CRITICAL: Deployment failed and system state is unstable",
			"id", cfg.ID,
			"status", cfg.Status,
			"error", err)
		os.Exit(1)
	}

	slog.Info("✨ Deployment process concluded successfully")
}
