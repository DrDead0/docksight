package installer

import "io"

// Reporter receives progress from an installation. The installer never
// imports the ui package: the cmd layer supplies an implementation, which
// keeps presentation out of the business logic and lets tests run silently.
type Reporter interface {
	// Step announces the start of a phase.
	Step(message string)

	// Success reports a completed phase.
	Success(message string)

	// Warn reports something the user should know but which is not fatal.
	Warn(message string)

	// Progress is where a subprocess writes its own output, such as docker
	// compose pull and startup logs.
	Progress() io.Writer
}

// Discard is a Reporter that swallows everything.
type Discard struct{}

func (Discard) Step(string)         {}
func (Discard) Success(string)      {}
func (Discard) Warn(string)         {}
func (Discard) Progress() io.Writer { return io.Discard }
