package deploy

import "context"

// Deployer defines the contract for infrastructure updates.
type Deployer interface {
	// Deploy pushes a specific version to a target environment
	Deploy(ctx context.Context, version string, weight int) error
	// GetRolloutStatus checks if the pods/instances are actually ready
	GetRolloutStatus(ctx context.Context) (bool, error)
}
