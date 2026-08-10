package installer

import (
	"context"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/envpath"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/filesystem"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/firewall"
)

// Host is the machine-wide configuration an install performs outside its own
// directories.
//
// It is an interface for the same reason Stack is, and one more. Everything
// behind it is a global side effect that no temporary directory contains: a
// machine PATH entry lives in HKLM and a firewall rule lives in the Windows
// Firewall, and both outlive the process that created them. Without a seam
// here the installer's own tests would write a real PATH entry pointing at a
// directory t.TempDir was about to delete.
type Host interface {
	// EnsureOnPath makes directory searchable, reporting whether it had to
	// change anything.
	EnsureOnPath(directory string) (bool, error)

	// AllowPort opens an inbound TCP port, reporting whether it had to.
	AllowPort(ctx context.Context, port int) (bool, error)

	// ProtectSecret restricts a file holding credentials to privileged
	// accounts.
	ProtectSecret(path string) error
}

// LocalHost applies the changes to the machine this runs on.
type LocalHost struct{}

func (LocalHost) EnsureOnPath(directory string) (bool, error) {
	return envpath.Ensure(directory)
}

func (LocalHost) AllowPort(ctx context.Context, port int) (bool, error) {
	return firewall.AllowPort(ctx, port)
}

// ProtectSecret is behind this seam and not called directly for a reason
// worth stating: on Windows it hands the file to Administrators and SYSTEM
// and takes it away from everyone else, including the unelevated half of the
// installing user's own token. That is correct for an installed platform and
// fatal for a test, which runs unelevated and would lock itself out of the
// .env it is checking. The real behaviour is asserted against the real
// implementation in the filesystem package instead.
func (LocalHost) ProtectSecret(path string) error {
	return filesystem.ProtectSecret(path)
}

// compile-time proof that the real implementation satisfies the interface.
var _ Host = LocalHost{}
