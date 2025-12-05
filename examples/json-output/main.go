package main

import (
	"context"
	"os"

	"github.com/deviantony/drillog"
)

func main() {
	// Use JSON output
	drillog.SetDefault(drillog.NewJSONHandler(os.Stderr, nil))

	ctx := context.Background()
	ctx, end := drillog.Start(ctx, "main")
	defer end()

	drillog.Info(ctx, "hello", "key", "value")
}
