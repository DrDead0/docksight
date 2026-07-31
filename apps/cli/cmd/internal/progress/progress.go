// Package progress defines how long-running operations report to the user.
//
// It exists so business-logic packages can describe what they are doing
// without importing the ui package: the cmd layer supplies the implementation.
package progress

import "io"

// Reporter receives progress from an operation.
type Reporter interface {
	// Step announces the start of a phase.
	Step(message string)

	// Success reports a completed phase.
	Success(message string)

	// Warn reports something the user should know but which is not fatal.
	Warn(message string)

	// Progress is where a subprocess writes its own output.
	Progress() io.Writer
}

// Discard is a Reporter that swallows everything, for tests and non
// interactive use.
type Discard struct{}

func (Discard) Step(string)         {}
func (Discard) Success(string)      {}
func (Discard) Warn(string)         {}
func (Discard) Progress() io.Writer { return io.Discard }
