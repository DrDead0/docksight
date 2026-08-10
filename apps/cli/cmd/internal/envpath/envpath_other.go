//go:build !windows

package envpath

// Ensure does nothing off Windows.
//
// The CLI installs into /usr/local/bin, which is on the default PATH of every
// shell on every distribution DockSight supports. Editing a shell profile to
// add a directory that is already searched would be a change with no effect
// and a file to clean up later.
func Ensure(string) (bool, error) {
	return false, nil
}
