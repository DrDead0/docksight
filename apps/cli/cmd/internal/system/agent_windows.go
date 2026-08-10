//go:build windows

package system

import (
	"context"
	"fmt"

	"golang.org/x/sys/windows"
)

// agentServiceRequirements is what supervising the agent needs on Windows.
//
// systemd has no counterpart here, so neither does the check for it. Two
// checks replace it, because the install genuinely depends on two things: a
// Service Control Manager that answers, and a process allowed to write to it.
//
// The elevation check belongs in validation rather than at the point of use.
// By the time the SCM refuses a CreateService call the binary has already
// been copied into Program Files and the configuration written, and all the
// operator sees is "Access is denied." from deep inside a syscall.
func agentServiceRequirements() []Requirement {

	return []Requirement{
		{
			Name:  "Checking the Service Control Manager",
			Check: func(context.Context) error { return CheckServiceManager() },
		},
		{
			Name:  "Checking Administrator rights",
			Check: func(context.Context) error { return CheckElevation() },
		},
	}
}

// CheckServiceManager reports whether the Service Control Manager answers.
//
// The handle is opened with SC_MANAGER_CONNECT, not SC_MANAGER_ALL_ACCESS:
// this check must fail only when the SCM itself is unreachable. Asking for
// write access would make an ordinary unelevated shell look like a broken
// host, and report it under the wrong name — elevation is the next check.
func CheckServiceManager() error {

	handle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)

	if err != nil {
		return fmt.Errorf("cannot reach the Service Control Manager: %w", err)
	}

	defer windows.CloseServiceHandle(handle)

	return nil
}

// CheckElevation reports whether this process may install a system service.
//
// UAC splits an administrator's token in two, and the unelevated half is a
// member of the Administrators group while being denied everything the group
// grants. Asking the token whether it is elevated is therefore the question
// that matters; group membership is not.
func CheckElevation() error {

	if windows.GetCurrentProcessToken().IsElevated() {
		return nil
	}

	return &NotElevatedError{
		Reason: "this process is not elevated, and registering a Windows service needs Administrator rights",
	}
}

// ElevationHint is how an operator obtains the rights an install needs.
func ElevationHint() string {
	return "Run this command from an Administrator prompt"
}
