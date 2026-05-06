package mock

import (
	"context"
	"log/slog"
	"time"
)

type Deployer struct {
	SimulateLatency time.Duration
}

func (m *Deployer) Deploy(ctx context.Context, version string, weight int) error {
	slog.Info("[MOCK] Simulating infrastructure update",
		"version", version,
		"traffic_weight", weight)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.SimulateLatency):
		return nil
	}
}
