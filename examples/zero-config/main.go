package main

import (
	"context"

	"github.com/deviantony/drillog"
)

func main() {
	// No setup - uses slog.Default()
	ctx := context.Background()
	ctx, end := drillog.Start(ctx, "main")
	defer end()

	drillog.Info(ctx, "zero config works", "test", true)

	nested(ctx)
}

func nested(ctx context.Context) {
	ctx, end := drillog.Start(ctx, "nested")
	defer end()

	drillog.Info(ctx, "inside nested span")
}
