package compose

import (
	"context"
	"io"
	"time"
)

// Runner drives one compose project. It bundles the polling interval so
// callers do not have to repeat it, and gives the installer a value it can
// substitute in tests.
type Runner struct {
	// ProjectDir is the directory compose runs from. Relative paths inside
	// the compose file and the project's .env resolve against it.
	ProjectDir string

	// File is the compose file name, relative to ProjectDir.
	File string

	// Interval is how often the stack is polled while waiting.
	Interval time.Duration
}

// NewRunner returns a runner for a compose project.
func NewRunner(projectDir string, file string) *Runner {

	return &Runner{
		ProjectDir: projectDir,
		File:       file,
		Interval:   2 * time.Second,
	}
}

// Start brings the stack up in the background.
func (r *Runner) Start(ctx context.Context, progress io.Writer) error {
	return Up(ctx, r.ProjectDir, r.File, progress)
}

// Stop takes the stack down.
func (r *Runner) Stop(ctx context.Context, progress io.Writer) error {
	return Down(ctx, r.ProjectDir, r.File, progress)
}

// Restart recreates the containers, which is what a platform update needs
// once new compose content has landed on disk.
func (r *Runner) Restart(ctx context.Context, progress io.Writer) error {

	if err := Down(ctx, r.ProjectDir, r.File, progress); err != nil {
		return err
	}

	return Up(ctx, r.ProjectDir, r.File, progress)
}

// WaitReady blocks until every service is ready or ctx is done.
func (r *Runner) WaitReady(ctx context.Context) ([]Service, error) {
	return WaitReady(ctx, r.ProjectDir, r.File, r.Interval)
}

// Status returns the current state of every service.
func (r *Runner) Status(ctx context.Context) ([]Service, error) {
	return Status(ctx, r.ProjectDir, r.File)
}
