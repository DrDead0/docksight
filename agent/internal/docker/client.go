package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// Client talks to Docker Engine over the local socket / named pipe.
type Client struct {
	http   *http.Client
	socket string
}

// NewClient connects to Docker Engine using the configured socket/named pipe.
func NewClient(socket string) (*Client, error) {
	if socket == "" {
		socket = DefaultSocket()
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialDocker(ctx, socket)
		},
	}

	return &Client{
		socket: socket,
		http: &http.Client{
			Transport: transport,
			Timeout:   20 * time.Second,
		},
	}, nil
}

// DefaultSocket returns the platform Docker Engine endpoint path.
func DefaultSocket() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\docker_engine`
	}
	return "/var/run/docker.sock"
}

func dialDocker(ctx context.Context, socket string) (net.Conn, error) {
	if runtime.GOOS == "windows" {
		return winio.DialPipeContext(ctx, socket)
	}
	d := net.Dialer{Timeout: 5 * time.Second}
	return d.DialContext(ctx, "unix", socket)
}

// Close releases HTTP resources.
func (c *Client) Close() error {
	c.http.CloseIdleConnections()
	return nil
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

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker"+path, nil)
	if err != nil {
		return err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("docker api %s: status %d: %s", path, res.StatusCode, strings.TrimSpace(string(body)))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}
