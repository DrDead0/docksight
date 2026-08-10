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

// ValidateAll is the diagnostic counterpart to Validate: it must run every
// check even after one fails, so `docksight doctor` can list all problems.
func TestValidateAllContinuesPastFailures(t *testing.T) {

	ran := 0

	requirements := []Requirement{
		{Name: "first", Check: func(context.Context) error {
			ran++
			return errors.New("first failed")
		}},
		{Name: "second", Check: func(context.Context) error {
			ran++
			return nil
		}},
		{Name: "third", Check: func(context.Context) error {
			ran++
			return errors.New("third failed")
		}},
	}

	record := &reporter{}
	err := ValidateAll(context.Background(), record, requirements)

	if err == nil {
		t.Fatal("expected a non-nil error when checks fail")
	}

	if ran != 3 {
		t.Fatalf("%d requirements ran, want 3", ran)
	}

	if !strings.Contains(err.Error(), "2 check(s) failed") {
		t.Fatalf("error summary wrong: %v", err)
	}

	if len(record.steps) != 3 {
		t.Fatalf("steps not numbered for every check: %v", record.steps)
	}
}

func TestValidateAllSucceedsWhenAllPass(t *testing.T) {

	requirements := []Requirement{
		{Name: "ok", Check: func(context.Context) error { return nil }},
	}

	if err := ValidateAll(context.Background(), &reporter{}, requirements); err != nil {
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

	joined := strings.ToLower(strings.Join(names, "|"))

	// Outbound connectivity is needed on every platform: the agent binary
	// comes from GitHub.
	if !strings.Contains(joined, "internet") {
		t.Errorf("agent requirements do not check connectivity: %v", names)
	}

	// The service manager is platform-specific, and checking the wrong one
	// would reject a host that can run the agent perfectly well.
	required := "systemd"
	forbidden := "service control manager"

	if runtime.GOOS == "windows" {
		required, forbidden = "service control manager", "systemd"
	}

	if !strings.Contains(joined, required) {
		t.Errorf("agent requirements do not check %s: %v", required, names)
	}

	if strings.Contains(joined, forbidden) {
		t.Errorf("agent requirements check %s on %s: %v", forbidden, runtime.GOOS, names)
	}
}

// Elevation is checked up front on Windows because the alternative is a raw
// "Access is denied." from the SCM after the binary has already been copied.
func TestAgentRequirementsCheckElevationOnWindows(t *testing.T) {

	if runtime.GOOS != "windows" {
		t.Skip("elevation is surfaced as a permission error on this platform")
	}

	for _, requirement := range AgentRequirements() {

		if strings.Contains(strings.ToLower(requirement.Name), "administrator") {
			return
		}
	}

	t.Fatal("agent requirements do not check for Administrator rights")
}

func TestCheckOS(t *testing.T) {

	err := CheckOS()

	switch runtime.GOOS {

	case "linux", "windows":

		if err != nil {
			t.Fatalf("%s rejected: %v", runtime.GOOS, err)
		}

	default:

		if err == nil {
			t.Fatalf("%s was accepted", runtime.GOOS)
		}

		if !strings.Contains(err.Error(), runtime.GOOS) {
			t.Fatalf("error does not name the OS: %v", err)
		}
	}
}

// Windows needs two checks Linux does not, and both exist to turn a failure
// that happens late and opaquely into one that happens first and says what
// to do.
func TestPlatformRequirementsOnWindows(t *testing.T) {

	if runtime.GOOS != "windows" {

		if len(platformExtraRequirements()) != 0 {
			t.Fatal("a non-Windows host contributed Windows-only platform checks")
		}

		return
	}

	names := make([]string, 0)

	for _, requirement := range PlatformRequirements() {
		names = append(names, strings.ToLower(requirement.Name))
	}

	joined := strings.Join(names, "|")

	for _, required := range []string{"linux container", "administrator"} {

		if !strings.Contains(joined, required) {
			t.Errorf("platform requirements do not check %q: %v", required, names)
		}
	}
}

func TestElevationHintNamesThisPlatform(t *testing.T) {

	hint := ElevationHint()

	want := "root"

	if runtime.GOOS == "windows" {
		want = "Administrator"
	}

	if !strings.Contains(hint, want) {
		t.Fatalf("hint %q does not mention %s", hint, want)
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
