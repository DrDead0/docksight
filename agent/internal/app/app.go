package app

import (
	"fmt"
	"os"
	"runtime"

	"docksight-agent/internal/communication"
	"docksight-agent/internal/config"
	"docksight-agent/internal/identity"
	"docksight-agent/internal/lifecycle"
	"docksight-agent/internal/logger"
	"docksight-agent/internal/version"
)

// App is the DockSight agent application bootstrapper.
type App struct {
	configPath string
}

// New creates an application instance.
func New(configPath string) *App {
	if configPath == "" {
		configPath = "config.yaml"
	}
	return &App{configPath: configPath}
}

// Run loads configuration, establishes identity, connects to the server,
// and blocks until interrupted.
func (a *App) Run() error {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	logger.Setup(cfg.Logging.Level, os.Stdout)
	logger.Info("configuration loaded", "path", a.configPath, "server", cfg.Server.URL)

	id, created, err := identity.LoadOrCreate(cfg.Agent.IdentityFile)
	if err != nil {
		return fmt.Errorf("identity: %w", err)
	}
	if created {
		logger.Info("identity created", "id", id.ID, "path", cfg.Agent.IdentityFile)
	} else {
		logger.Info("identity loaded", "id", id.ID, "path", cfg.Agent.IdentityFile)
	}

	dockerSocketFound := socketExists(cfg.Docker.Socket)
	if dockerSocketFound {
		logger.Info("docker socket found", "socket", cfg.Docker.Socket)
	} else {
		logger.Warn("docker socket not found", "socket", cfg.Docker.Socket)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	printStartupSummary(id.ID, created, dockerSocketFound, cfg.Server.URL)

	lc := lifecycle.New()
	client := communication.NewClient(cfg.Server.URL, communication.AgentInfo{
		UUID:         id.ID,
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Version:      version.Version,
	})

	go client.Run(lc.Context())

	logger.Info("agent ready; waiting for shutdown signal",
		"version", version.Version,
		"commit", version.Commit,
		"buildDate", version.BuildDate,
		"agentId", id.ID,
	)
	lc.Wait(func() {
		logger.Info("agent stopped")
	})

	return nil
}

func printStartupSummary(agentID string, identityCreated bool, dockerSocketFound bool, serverURL string) {
	logger.Printf("DockSight Agent %s\n\n", version.String())
	logger.Printf("Configuration loaded\n")
	if identityCreated {
		logger.Printf("Identity created (%s)\n", agentID)
	} else {
		logger.Printf("Identity loaded (%s)\n", agentID)
	}
	if dockerSocketFound {
		logger.Printf("Docker socket found\n")
	} else {
		logger.Printf("Docker socket not found\n")
	}
	logger.Printf("Server: %s\n", serverURL)
	logger.Printf("\nAgent Status: READY\n\n")
}

func socketExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
