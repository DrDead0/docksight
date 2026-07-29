package compose

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Service is one entry of `docker compose ps --format json`.
type Service struct {
	Name     string `json:"Name"`
	Service  string `json:"Service"`
	State    string `json:"State"`
	Health   string `json:"Health"`
	ExitCode int    `json:"ExitCode"`
}

// Ready reports whether the service is up. A service that declares a
// healthcheck must also report healthy — "running" only means the process
// started, not that it can serve traffic.
func (s Service) Ready() bool {
	return s.State == "running" && (s.Health == "" || s.Health == "healthy")
}

// Failed reports whether the service has given up: a container that exited
// non-zero or died is never going to become ready on its own.
func (s Service) Failed() bool {

	if s.State == "dead" {
		return true
	}

	return s.State == "exited" && s.ExitCode != 0
}

// Describe renders the service state for a human, e.g. "postgres: running
// (unhealthy)".
func (s Service) Describe() string {

	switch {

	case s.State == "exited":
		return fmt.Sprintf("%s: exited (code %d)", s.Service, s.ExitCode)

	case s.Health != "":
		return fmt.Sprintf("%s: %s (%s)", s.Service, s.State, s.Health)

	default:
		return fmt.Sprintf("%s: %s", s.Service, s.State)
	}
}

// Up starts the stack in the background, streaming docker's own progress
// output to progress. Pass nil to discard it.
func Up(ctx context.Context, projectDir string, composeFile string, progress io.Writer) error {

	if progress == nil {
		progress = io.Discard
	}

	command := docker(ctx, projectDir, composeFile, "up", "--detach", "--remove-orphans")

	command.Stdout = progress
	command.Stderr = progress

	if err := command.Run(); err != nil {
		return fmt.Errorf("docker compose up failed: %w", err)
	}

	return nil
}

// Down stops the stack and removes its containers.
func Down(ctx context.Context, projectDir string, composeFile string, progress io.Writer) error {

	if progress == nil {
		progress = io.Discard
	}

	command := docker(ctx, projectDir, composeFile, "down", "--remove-orphans")

	command.Stdout = progress
	command.Stderr = progress

	if err := command.Run(); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	return nil
}

// Status returns the current state of every container in the project,
// including ones that have exited.
func Status(ctx context.Context, projectDir string, composeFile string) ([]Service, error) {

	command := docker(ctx, projectDir, composeFile, "ps", "--all", "--format", "json")

	var stdout, stderr bytes.Buffer

	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf(
			"docker compose ps failed: %w: %s",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}

	return parseServices(stdout.Bytes())
}

// WaitReady polls the project until every service is ready, returning the
// final service list. It gives up early if a container has exited for good,
// and otherwise stops when ctx is done.
func WaitReady(
	ctx context.Context,
	projectDir string,
	composeFile string,
	interval time.Duration,
) ([]Service, error) {

	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var last []Service

	for {

		services, err := Status(ctx, projectDir, composeFile)

		if err != nil {
			return last, err
		}

		last = services

		if len(services) == 0 {
			return last, fmt.Errorf("no services found in %s", composeFile)
		}

		if failed := filter(services, Service.Failed); len(failed) > 0 {
			return last, fmt.Errorf(
				"service failed to start: %s (inspect with: docker compose -f %s logs)",
				describe(failed),
				composeFile,
			)
		}

		pending := filter(services, func(s Service) bool { return !s.Ready() })

		if len(pending) == 0 {
			return last, nil
		}

		select {

		case <-ctx.Done():
			return last, fmt.Errorf(
				"timed out waiting for services: %s",
				describe(pending),
			)

		case <-ticker.C:
		}
	}
}

// docker builds a `docker compose` invocation rooted at projectDir, so that
// relative paths inside the compose file and the project's .env resolve.
func docker(
	ctx context.Context,
	projectDir string,
	composeFile string,
	args ...string,
) *exec.Cmd {

	command := exec.CommandContext(
		ctx,
		"docker",
		append([]string{"compose", "--file", composeFile}, args...)...,
	)

	command.Dir = projectDir

	return command
}

// parseServices accepts both shapes docker has emitted for `ps --format
// json`: a single JSON array, and one JSON object per line.
func parseServices(output []byte) ([]Service, error) {

	trimmed := bytes.TrimSpace(output)

	if len(trimmed) == 0 {
		return nil, nil
	}

	if trimmed[0] == '[' {

		var services []Service

		if err := json.Unmarshal(trimmed, &services); err != nil {
			return nil, fmt.Errorf("failed to parse docker compose ps output: %w", err)
		}

		return services, nil
	}

	var services []Service

	decoder := json.NewDecoder(bytes.NewReader(trimmed))

	for {

		var service Service

		err := decoder.Decode(&service)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse docker compose ps output: %w", err)
		}

		services = append(services, service)
	}

	return services, nil
}

func filter(services []Service, keep func(Service) bool) []Service {

	var kept []Service

	for _, service := range services {
		if keep(service) {
			kept = append(kept, service)
		}
	}

	return kept
}

func describe(services []Service) string {

	descriptions := make([]string, 0, len(services))

	for _, service := range services {
		descriptions = append(descriptions, service.Describe())
	}

	return strings.Join(descriptions, ", ")
}
