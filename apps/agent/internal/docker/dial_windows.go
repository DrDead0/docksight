//go:build windows

package docker

import (
	"context"
	"net"

	"github.com/Microsoft/go-winio"
)

// dialDocker connects to the Docker Engine named pipe.
//
// This lives in a Windows-only file because go-winio does not build for any
// other platform: importing it unconditionally makes the agent impossible to
// cross-compile for the Linux hosts it actually runs on.
func dialDocker(ctx context.Context, socket string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, socket)
}
