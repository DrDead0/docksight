package main

import (
	"fmt"
	"os"

	"docksight-agent/internal/config"
)

// DockSight agent entrypoint.
// Docker Engine API integration and backend WebSocket communication
// will be implemented in later milestones.
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("DockSight agent foundation ready (server=%s)\n", cfg.ServerURL)
}
