package cmd

import (
	"io"
	"os"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

// consoleReporter renders installer progress. It is the only bridge between
// the installer and the ui package, which is what keeps presentation out of
// the internal packages.
type consoleReporter struct{}

func (consoleReporter) Step(message string) { ui.Info(message) }

func (consoleReporter) Success(message string) { ui.Success(message) }

func (consoleReporter) Warn(message string) { ui.Warning(message) }

// Progress sends subprocess output to stdout, the same stream the ui package
// writes to, so `docksight install > install.log` captures the whole run.
func (consoleReporter) Progress() io.Writer { return os.Stdout }
