//go:build !windows

package system

import (
	"context"
	"os"
)

// platformExtraRequirements is what `docksight install` needs beyond Docker
// on this platform. On Linux that is nothing: root is surfaced as a
// permission error when the install writes, which is the behaviour operators
// already know, and the Engine has no container mode to get wrong.
func platformExtraRequirements() []Requirement {
	return nil
}

// agentServiceRequirements is what supervising the agent needs on this
// platform.
//
// On Linux that is systemd and nothing else. Elevation is deliberately not
// checked here: the install has always surfaced a missing root as a
// permission error carrying the "Run this command as root" hint, and turning
// that into a preflight failure would change a message operators already
// recognise. Windows has no such established behaviour, which is why it
// checks elevation up front instead.
func agentServiceRequirements() []Requirement {

	return []Requirement{
		{Name: "Checking systemd", Check: func(context.Context) error { return CheckSystemd() }},
	}
}

// CheckElevation reports whether this process may install a system service.
func CheckElevation() error {

	if os.Geteuid() == 0 {
		return nil
	}

	return &NotElevatedError{Reason: "this command needs root privileges"}
}

// ElevationHint is how an operator obtains the rights an install needs.
func ElevationHint() string {
	return "Run this command as root"
}
