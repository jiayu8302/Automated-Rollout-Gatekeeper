package azure

import (
	"context"
	"log/slog"
)

// Deployer must be CAPITALIZED to be exported
type Deployer struct {
	ResourceGroup string
	AppName       string
}

func (a *Deployer) Deploy(ctx context.Context, version string, weight int) error {
	slog.Info("[Azure] Updating Container App",
		"resource_group", a.ResourceGroup,
		"app", a.AppName,
		"version", version,
		"weight", weight)
	return nil
}
