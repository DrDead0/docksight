// Package system validates that a host can run DockSight.
//
// Every check returns a real error. A validation step that reports a problem
// and then continues is worse than no check at all: the install proceeds,
// writes files, and fails later with an unrelated message.
package system

import (
	"context"
	"fmt"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/progress"
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

// ValidateAll runs every requirement and reports each result. Unlike
// Validate, it does not stop at the first failure — useful for diagnostic
// commands that should surface every problem at once.
//
// The returned error is non-nil when any check failed. Individual failures
// are reported via reporter.Warn so the caller can still print a summary.
func ValidateAll(ctx context.Context, reporter progress.Reporter, requirements []Requirement) error {

	if reporter == nil {
		reporter = progress.Discard{}
	}

	total := len(requirements)
	failed := 0

	for index, requirement := range requirements {

		reporter.Step(fmt.Sprintf("[%d/%d] %s", index+1, total, requirement.Name))

		if err := requirement.Check(ctx); err != nil {
			failed++
			reporter.Warn(fmt.Sprintf("%s — %v", requirement.Name, err))
			continue
		}

		reporter.Success(requirement.Name)
	}

	if failed == 0 {
		return nil
	}

	return fmt.Errorf("%d check(s) failed", failed)
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

// AgentRequirements is what `docksight agent install` needs. The agent is
// supervised by the host's service manager and downloads its binary from
// GitHub, so it needs two things the platform install does not: a usable
// service manager, and outbound connectivity.
//
// The middle of the list is platform-specific. Checking systemd on a Windows
// host would stat a path that cannot exist and reject a machine that is
// perfectly able to run the agent, so each platform contributes its own
// checks through agentServiceRequirements.
func AgentRequirements() []Requirement {

	requirements := []Requirement{
		{Name: "Checking operating system", Check: func(context.Context) error { return CheckAgentOS() }},
		{Name: "Checking architecture", Check: func(context.Context) error { return CheckArchitecture() }},
		{Name: "Checking Docker Engine", Check: CheckDockerInstalled},
		{Name: "Checking Docker daemon", Check: CheckDockerRunning},
	}

	requirements = append(requirements, agentServiceRequirements()...)

	return append(requirements, Requirement{
		Name:  "Checking internet connectivity",
		Check: CheckInternet,
	})
}
