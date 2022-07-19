package main

import (
	"context"
	"os"
	"zombiezen.com/go/log"
)

const DaprSocketPathEnvVar = "DAPR_COMPONENT_SOCKET_PATH"

func main() {
	log.SetDefault(&log.LevelFilter{
		Min:    log.Debug, // Only show warnings or above
		Output: log.New(os.Stdout, "LOG: ", 0, nil),
	})
	ctx := context.Background()
	socketPath := os.Getenv(DaprSocketPathEnvVar)
	log.Debugf(ctx, "Using socket path: %s", socketPath)
	memstore := NewMemoryStore(ctx)
	memstore.Serve(socketPath)
}
