//go:build windows

package service

import (
	"context"

	"golang.org/x/sys/windows/svc"
)

// IsService reports whether the Service Control Manager started this process.
//
// The same executable must work both ways: a developer runs it in a console,
// the SCM runs it as a service. Detecting that at runtime is what removes the
// need for a --service flag that the SCM would have to be configured to pass
// and a developer would have to remember not to.
func IsService() bool {
	is, err := svc.IsWindowsService()
	return err == nil && is
}

// Run executes the agent, under the SCM when the SCM started us.
func Run(run func(context.Context) error) error {

	if !IsService() {
		return run(context.Background())
	}

	return svc.Run(Name, &handler{run: run})
}

// handler translates Service Control Manager control codes into the context
// cancellation the agent already shuts down on.
type handler struct {
	run func(context.Context) error
}

// Execute is the SCM's entry point, and is required to keep the status channel
// current: the SCM decides whether a service is healthy, starting or hung
// purely from what arrives here.
func (h *handler) Execute(
	_ []string,
	requests <-chan svc.ChangeRequest,
	status chan<- svc.Status,
) (bool, uint32) {

	status <- svc.Status{State: svc.StartPending}

	events := openEventLog()
	defer events.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	finished := make(chan error, 1)

	go func() { finished <- h.run(ctx) }()

	// AcceptShutdown is listed separately from AcceptStop on purpose. Without
	// it a machine reboot terminates the agent instead of stopping it, and the
	// shutdown hooks never run.
	status <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

	events.info("DockSight Agent started")

	for {
		select {

		case request := <-requests:

			switch request.Cmd {

			case svc.Interrogate:
				// A health poll. Failing to answer makes the SCM consider the
				// service unresponsive.
				status <- request.CurrentStatus

			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending}

				// The same cancellation Ctrl+C produces, so the existing
				// shutdown hooks close the log streams and Docker client.
				cancel()

				if err := <-finished; err != nil {
					events.error("DockSight Agent stopped with an error: " + err.Error())
					return false, 1
				}

				events.info("DockSight Agent stopped")

				return false, 0

			default:
				// Control codes we never asked to accept. Ignoring them is
				// correct; reporting an error would fail the service.
			}

		case err := <-finished:

			// The agent exited on its own — a bad config, an unusable Docker
			// endpoint. Reporting a non-zero code lets configured recovery
			// actions see a failure rather than a clean stop.
			status <- svc.Status{State: svc.StopPending}

			if err != nil {
				events.error("DockSight Agent exited: " + err.Error())
				return false, 1
			}

			events.info("DockSight Agent exited")

			return false, 0
		}
	}
}
