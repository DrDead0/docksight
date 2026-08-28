package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"docksight-agent/internal/app"
	"docksight-agent/internal/service"
	"docksight-agent/internal/version"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to agent config.yaml")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	application := app.New(*configPath)

	// service.Run decides how the agent is supervised: directly in a console,
	// or dispatched to the Windows Service Control Manager. The same binary
	// serves both, detected at runtime rather than selected by a flag.
	err := service.Run(func(ctx context.Context) error {
		return application.Run(ctx)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "docksight-agent: %v\n", err)
		os.Exit(1)
	}
}
