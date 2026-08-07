package main

import (
	"flag"
	"fmt"
	"os"

	"docksight-agent/internal/app"
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

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "docksight-agent: %v\n", err)
		os.Exit(1)
	}
}
