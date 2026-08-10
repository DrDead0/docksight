//go:build !windows

package firewall

import "context"

// AllowPort does nothing off Windows. See the package comment for why a Linux
// host's firewall is left to its administrator.
func AllowPort(context.Context, int) (bool, error) {
	return false, nil
}
