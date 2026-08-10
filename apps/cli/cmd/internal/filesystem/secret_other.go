//go:build !windows

package filesystem

import (
	"fmt"
	"os"
)

// ProtectSecret restricts a file to its owner.
//
// The mode is reasserted rather than assumed. env.Generate creates the file
// at 0600 already, but a file carried over from an earlier install, or one an
// operator has edited and saved with a permissive umask, is exactly the case
// worth repairing — and doing it here means neither caller has to know which
// platform needs what.
func ProtectSecret(path string) error {

	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("failed to restrict access to %s: %w", path, err)
	}

	return nil
}
