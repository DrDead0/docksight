package main

import (
	"flag"
	"fmt"
	"os"

	"docksight-agent/internal/app"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to agent config.yaml")
	flag.Parse()
	application := app.New(*configPath)
	
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "docksight-agent: %v\n", err)
		os.Exit(1)
	}
}
