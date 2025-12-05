package main

import (
	"context"
	"os"
	"time"

	"github.com/deviantony/drillog"
)

func main() {
	// Configure handler
	drillog.SetDefault(drillog.NewTextHandler(os.Stderr, nil))

	ctx := context.Background()
	ctx, end := drillog.Start(ctx, "sync-cycle")
	defer end()

	drillog.Info(ctx, "starting device sync")

	// Fetch agents
	fetchAgents(ctx)

	// Process devices
	processDevice(ctx, "device-001")
	processDevice(ctx, "device-002")
	processDevice(ctx, "device-003")

	drillog.Info(ctx, "sync complete", "devices", 3)
}

func fetchAgents(ctx context.Context) {
	ctx, end := drillog.Start(ctx, "fetch-agents")
	defer end()

	drillog.Debug(ctx, "connecting to API")
	time.Sleep(10 * time.Millisecond)
	drillog.Info(ctx, "agents retrieved", "count", 5)
}

func processDevice(ctx context.Context, deviceID string) {
	ctx, end := drillog.Start(ctx, "process-device")
	defer end()

	drillog.Info(ctx, "processing", "device_id", deviceID)
	time.Sleep(5 * time.Millisecond)

	// Nested operation
	syncData(ctx, deviceID)

	drillog.Info(ctx, "device processed", "device_id", deviceID)
}

func syncData(ctx context.Context, deviceID string) {
	ctx, end := drillog.Start(ctx, "sync-data")
	defer end()

	drillog.Debug(ctx, "syncing data", "device_id", deviceID)
	time.Sleep(2 * time.Millisecond)
}
