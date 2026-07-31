package install

import (
	"fmt"
	"strings"
)

// Installation phases, named so a failure says which one broke.
const (
	PhaseValidate  = "system validation"
	PhaseDownload  = "downloading the agent"
	PhaseBinary    = "installing the binary"
	PhaseConfigure = "creating the configuration"
	PhaseService   = "creating the systemd service"
	PhaseVerify    = "verifying the installation"
)

// PhaseError reports which phase failed, why, and what to do about it.
// Returning a bare error from six phases makes "permission denied" ambiguous;
// this makes the failure self-describing.
type PhaseError struct {
	// Phase is the installation phase that failed.
	Phase string

	// Err is the underlying failure.
	Err error

	// Hint is an actionable next step, when one is known.
	Hint string

	// Logs is service output captured at the point of failure.
	Logs string
}

func (e *PhaseError) Error() string {

	var message strings.Builder

	fmt.Fprintf(&message, "%s failed: %v", e.Phase, e.Err)

	if e.Hint != "" {
		fmt.Fprintf(&message, "\n  %s", e.Hint)
	}

	if trimmed := strings.TrimSpace(e.Logs); trimmed != "" {
		fmt.Fprintf(&message, "\n\n  Recent service output:\n%s", indent(trimmed))
	}

	return message.String()
}

func (e *PhaseError) Unwrap() error {
	return e.Err
}

// fail wraps an error with the phase it came from.
func fail(phase string, err error, hint string) error {

	if err == nil {
		return nil
	}

	var alreadyWrapped *PhaseError

	if asPhaseError(err, &alreadyWrapped) {
		return err
	}

	return &PhaseError{Phase: phase, Err: err, Hint: hint}
}

func indent(text string) string {

	lines := strings.Split(text, "\n")

	for index, line := range lines {
		lines[index] = "    " + line
	}

	return strings.Join(lines, "\n")
}
