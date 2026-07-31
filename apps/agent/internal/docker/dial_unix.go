//go:build !windows

package docker

import (
	"context"
	"net"
	"time"
)

// dialDocker connects to the Docker Engine Unix socket.
func dialDocker(ctx context.Context, socket string) (net.Conn, error) {

	dialer := net.Dialer{Timeout: 5 * time.Second}

	return dialer.DialContext(ctx, "unix", socket)
}
