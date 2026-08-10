//go:build windows

package service

import "golang.org/x/sys/windows/svc/eventlog"

// eventID is the identifier every entry is written under. DockSight does not
// classify its events, and Event Viewer requires some value.
const eventID = 1

// events writes service lifecycle transitions to the Windows Event Log.
//
// Only lifecycle belongs here — started, stopped, exited. The agent's own
// structured output goes to the log file instead: Event Viewer is where an
// administrator looks to find out whether the service is running, not to read
// a debug stream, and a chatty source is a source people turn off.
type events struct {
	log *eventlog.Log
}

// openEventLog connects to the Event Log source, or returns a sink that
// silently discards.
//
// Registering the source needs Administrator rights and is done once by
// "docksight agent install". An agent that has not been installed that way —
// or was installed before the source existed — must still start and run: a
// missing Event Log source is a reduced-visibility problem, not a reason to
// refuse to monitor Docker. The log file still receives everything.
func openEventLog() *events {

	log, err := eventlog.Open(Name)

	if err != nil {
		return &events{}
	}

	return &events{log: log}
}

func (e *events) info(message string) {

	if e.log == nil {
		return
	}

	_ = e.log.Info(eventID, message)
}

func (e *events) error(message string) {

	if e.log == nil {
		return
	}

	_ = e.log.Error(eventID, message)
}

func (e *events) Close() error {

	if e.log == nil {
		return nil
	}

	return e.log.Close()
}
