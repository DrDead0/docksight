//go:build !windows

package install

import "github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/systemd"

// NewController returns the service manager this host runs.
func NewController(Layout) UnitController {
	return systemd.NewManager()
}
