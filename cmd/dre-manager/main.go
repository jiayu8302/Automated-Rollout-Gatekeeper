package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jiayu8302/deployment-reliability-engine/pkg/api"
	"github.com/jiayu8302/deployment-reliability-engine/pkg/engine"
	"github.com/jiayu8302/deployment-reliability-engine/pkg/providers/azure"
	"github.com/jiayu8302/deployment-reliability-engine/pkg/providers/k8s"
	"github.com/jiayu8302/deployment-reliability-engine/pkg/providers/mock"
)

// Config represents the internal configuration structure, mapping the target provider.
type Config struct {
	TargetProvider string         `yaml:"target_provider"`
	Deployment     api.Deployment `yaml:",inline"`
}

// RealWorldProber connects DRE to telemetry sources (e.g., Prometheus or Datadog).
type RealWorldProber struct{}

func (p *RealWorldProber) Check(ctx context.Context, d *api.Deployment) error {
	slog.Info("Running health analysis", "deployment_id", d.ID, "stage", d.Status)
	// In a real scenario, this would call an external API to check error rates.
	return nil
}

// NewDeployer acts as the Factory Pattern to load the cloud-specific driver.
func NewDeployer(providerType string) (engine.Deployer, error) {
	switch providerType {
	case "azure":
		return &azure.Deployer{
			ResourceGroup: "prod-resilience-rg",
			AppName:       "billing-api",
		}, nil
	case "kubernetes":
		return &k8s.Deployer{
			Namespace: "production",
			Resource:  "deployments/api-service",
		}, nil
	case "mock":
		return &mock.Deployer{SimulateLatency: 500 * time.Millisecond}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", providerType)
	}
}

// loadConfig reads the deployment policy from the local file system.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	return &cfg, nil
}

func main() {
	// 1. Initialize Structured JSON Logging for production traceability.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("🚀 DRE: Deployment Reliability Engine Initializing...")

	// 2. Load and validate Configuration.
	cfg, err := loadConfig("configs/deploy_policy.yaml")
	if err != nil {
		slog.Error("Bootstrap failed: Configuration error", "error", err)
		os.Exit(1)
	}

	// 3. Initialize the Cloud-Specific Deployer.
	deployer, err := NewDeployer(cfg.TargetProvider)
	if err != nil {
		slog.Error("Provider initialization failed", "error", err)
		os.Exit(1)
	}

	// 4. Setup the Rollout Engine.
	dre := &engine.DeploymentEngine{
		Prober:   &RealWorldProber{},
		Deployer: deployer,
	}

	// 5. Lifecycle and Signal Management.
	// We allow up to 10 minutes for the entire rollout process.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Capture interrupt signals (Ctrl+C) for graceful cancellation.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Warn("Termination signal received. Aborting rollout...", "signal", sig)
		cancel()
	}()

	// 6. Execute the Rollout State Machine.
	slog.Info("Deployment cycle starting",
		"id", cfg.Deployment.ID,
		"provider", cfg.TargetProvider,
		"target_version", cfg.Deployment.Version)

	if err := dre.ExecuteRollout(ctx, &cfg.Deployment); err != nil {
		slog.Error("CRITICAL: Rollout failed",
			"id", cfg.Deployment.ID,
			"final_status", cfg.Deployment.Status,
			"error", err)

		// Exit with failure so CI/CD pipelines can detect the issue.
		os.Exit(1)
	}

	slog.Info("✨ Rollout sequence concluded successfully", "id", cfg.Deployment.ID)
}
