package system

import (
	"context"
	"errors"
	"io"
	"runtime"
	"strings"
	"testing"
)

type reporter struct{ steps []string }

func (r *reporter) Step(message string) { r.steps = append(r.steps, message) }
func (r *reporter) Success(string)      {}
func (r *reporter) Warn(string)         {}
func (r *reporter) Progress() io.Writer { return io.Discard }

func TestValidateRunsEveryRequirement(t *testing.T) {

	ran := 0

	requirements := []Requirement{
		{Name: "first", Check: func(context.Context) error { ran++; return nil }},
		{Name: "second", Check: func(context.Context) error { ran++; return nil }},
	}

	record := &reporter{}

	if err := Validate(context.Background(), record, requirements); err != nil {
		t.Fatal(err)
	}

	if ran != 2 {
		t.Fatalf("%d requirements ran, want 2", ran)
	}

	if len(record.steps) != 2 || !strings.HasPrefix(record.steps[0], "[1/2]") {
		t.Fatalf("steps not numbered: %v", record.steps)
	}
}

// A failed check must stop the run. The whole point of validation is to fail
// before anything is written to the host.
func TestValidateStopsAtFirstFailure(t *testing.T) {

	reached := false

	requirements := []Requirement{
		{Name: "docker", Check: func(context.Context) error { return errors.New("docker is not installed") }},
		{Name: "systemd", Check: func(context.Context) error { reached = true; return nil }},
	}

	err := Validate(context.Background(), &reporter{}, requirements)

	if err == nil {
		t.Fatal("a failing requirement did not stop validation")
	}

	if reached {
		t.Fatal("validation continued past a failure")
	}

	if !strings.Contains(err.Error(), "docker is not installed") {
		t.Fatalf("error lost the cause: %v", err)
	}
}

func TestValidateWithoutReporter(t *testing.T) {

	requirements := []Requirement{
		{Name: "noop", Check: func(context.Context) error { return nil }},
	}

	if err := Validate(context.Background(), nil, requirements); err != nil {
		t.Fatal(err)
	}
}

func TestRequirementSets(t *testing.T) {

	if len(PlatformRequirements()) == 0 {
		t.Fatal("the platform has no requirements")
	}

	agent := AgentRequirements()

	names := make([]string, 0, len(agent))

	for _, requirement := range agent {
		names = append(names, requirement.Name)
	}

	joined := strings.Join(names, "|")

	// The two checks the agent needs and the platform install does not.
	for _, required := range []string{"systemd", "internet"} {

		if !strings.Contains(strings.ToLower(joined), required) {
			t.Errorf("agent requirements do not check %s: %v", required, names)
		}
	}
}

func TestCheckOS(t *testing.T) {

	err := CheckOS()

	if runtime.GOOS == "linux" && err != nil {
		t.Fatalf("linux rejected: %v", err)
	}

	if runtime.GOOS != "linux" {

		if err == nil {
			t.Fatal("a non-linux host was accepted")
		}

		if !strings.Contains(err.Error(), runtime.GOOS) {
			t.Fatalf("error does not name the OS: %v", err)
		}
	}
}

func TestCheckArchitecture(t *testing.T) {

	err := CheckArchitecture()

	switch runtime.GOARCH {

	case "amd64", "arm64":

		if err != nil {
			t.Fatalf("%s rejected: %v", runtime.GOARCH, err)
		}

	default:

		if err == nil {
			t.Fatalf("unsupported architecture %s accepted", runtime.GOARCH)
		}
	}
}
