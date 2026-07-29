package docker

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/docker/docker/client"
)

// Client wraps the Docker Engine Go SDK with a platform-aware dialer.
type Client struct {
	sdk    *client.Client
	socket string
}

// NewClient connects to Docker Engine using the configured socket/named pipe.
func NewClient(socket string) (*Client, error) {
	if socket == "" {
		socket = DefaultSocket()
	}

	opts := []client.Opt{
		client.WithHost(engineHost(socket)),
		client.WithAPIVersionNegotiation(),
		// Apply after WithHost so we override any transport dialer ConfigureTransport set.
		client.WithDialContext(func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialDocker(ctx, socket)
		}),
	}

	sdk, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("docker sdk client: %w", err)
	}

	return &Client{
		sdk:    sdk,
		socket: socket,
	}, nil
}

// SDK returns the underlying Docker Engine API client.
func (c *Client) SDK() *client.Client {
	return c.sdk
}

// DefaultSocket returns the platform Docker Engine endpoint path.
func DefaultSocket() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\docker_engine`
	}
	return "/var/run/docker.sock"
}

func engineHost(socket string) string {
	if runtime.GOOS == "windows" {
		if socket == "" || socket == `\\.\pipe\docker_engine` {
			return client.DefaultDockerHost
		}
		return "npipe://" + socket
	}
	if socket == "" {
		socket = "/var/run/docker.sock"
	}
	return "unix://" + socket
}

func dialDocker(ctx context.Context, socket string) (net.Conn, error) {
	if runtime.GOOS == "windows" {
		return winio.DialPipeContext(ctx, socket)
	}
	d := net.Dialer{Timeout: 5 * time.Second}
	return d.DialContext(ctx, "unix", socket)
}

// Close releases SDK resources.
func (c *Client) Close() error {
	if c.sdk == nil {
		return nil
	}
	return c.sdk.Close()
}

// Socket returns the configured engine socket path.
func (c *Client) Socket() string {
	return c.socket
}

// SocketExists reports whether the configured Docker socket/pipe path exists.
func SocketExists(socket string) bool {
	if socket == "" {
		socket = DefaultSocket()
	}
	_, err := os.Stat(socket)
	return err == nil
}
