//go:build windows

package install

import "github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/scm"

// compile-time proof that the SCM manager satisfies the interface. This is
// the only file that can import the package: scm builds against
// golang.org/x/sys/windows/svc, which exists nowhere else.
var _ UnitController = (*scm.Manager)(nil)

// NewController returns the service manager this host runs.
//
// The log path travels with the layout because the two must agree: the
// installer tells the agent where to write through config.yaml, and the
// controller reads the same file back to verify the agent connected. On Linux
// neither side needs telling — the agent writes to stdout and systemd owns
// the journal.
func NewController(layout Layout) UnitController {

	manager := scm.NewManager()
	manager.LogPath = layout.LogPath

	return manager
}
