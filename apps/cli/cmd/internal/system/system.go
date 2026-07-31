// Package system validates that a host can run DockSight.
//
// Every check returns a real error. A validation step that reports a problem
// and then continues is worse than no check at all: the install proceeds,
// writes files, and fails later with an unrelated message.
package system

import (
	"context"
	"fmt"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/progress"
)

// Requirement is one named precondition.
type Requirement struct {
	Name  string
	Check func(context.Context) error
}

// Validate runs requirements in order and stops at the first failure.
func Validate(ctx context.Context, reporter progress.Reporter, requirements []Requirement) error {

	if reporter == nil {
		reporter = progress.Discard{}
	}

	total := len(requirements)

	for index, requirement := range requirements {

		reporter.Step(fmt.Sprintf("[%d/%d] %s", index+1, total, requirement.Name))

		if err := requirement.Check(ctx); err != nil {
			return err
		}

		reporter.Success(requirement.Name)
	}

	return nil
}

// PlatformRequirements is what `docksight install` needs: a supported host
// with a working Docker Engine and Compose.
func PlatformRequirements() []Requirement {

	return []Requirement{
		{Name: "Checking operating system", Check: func(context.Context) error { return CheckOS() }},
		{Name: "Checking architecture", Check: func(context.Context) error { return CheckArchitecture() }},
		{Name: "Checking Docker Engine", Check: CheckDockerInstalled},
		{Name: "Checking Docker daemon", Check: CheckDockerRunning},
		{Name: "Checking Docker Compose", Check: CheckDockerCompose},
	}
}

// AgentRequirements is what `docksight agent install` needs. The agent runs
// as a systemd service and downloads its binary from GitHub, so it needs two
// things the platform install does not: systemd, and outbound connectivity.
func AgentRequirements() []Requirement {

	return []Requirement{
		{Name: "Checking operating system", Check: func(context.Context) error { return CheckOS() }},
		{Name: "Checking architecture", Check: func(context.Context) error { return CheckArchitecture() }},
		{Name: "Checking Docker Engine", Check: CheckDockerInstalled},
		{Name: "Checking Docker daemon", Check: CheckDockerRunning},
		{Name: "Checking systemd", Check: func(context.Context) error { return CheckSystemd() }},
		{Name: "Checking internet connectivity", Check: CheckInternet},
	}
}
