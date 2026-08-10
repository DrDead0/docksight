// Package service runs the agent under whatever supervises it.
//
// On Linux the agent is a systemd unit: it runs in the foreground, writes to
// stdout and stops on SIGTERM. On Windows the Service Control Manager starts
// the same executable with no console and stops it with a control code that
// bears no relation to a Unix signal.
//
// Run hides that difference behind one call so the rest of the agent is
// written once. The platform-specific halves live in run_windows.go and
// run_other.go, following the build-tag split established by the Docker
// dialler — golang.org/x/sys/windows/svc does not compile off Windows, and an
// unconditional import would break the Linux builds that matter most.
package service

// Name is how the Service Control Manager identifies the agent, and the
// Event Log source it writes lifecycle events to. The CLI registers both
// under this name when it installs the service, so the two must agree.
const Name = "docksight-agent"
