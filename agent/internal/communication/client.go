package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"docksight-agent/internal/logger"

	"github.com/gorilla/websocket"
)

const (
	TypeRegister   = "agent.register"
	TypeRegistered = "agent.registered"
	TypeHeartbeat  = "agent.heartbeat"
)

// Envelope is the JSON message wrapper exchanged with the DockSight server.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RegisterPayload is sent on agent.register.
type RegisterPayload struct {
	UUID         string `json:"uuid"`
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Version      string `json:"version"`
}

// RegisteredPayload is returned on agent.registered.
type RegisteredPayload struct {
	ID      string `json:"id"`
	UUID    string `json:"uuid"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HeartbeatPayload is sent on agent.heartbeat.
type HeartbeatPayload struct {
	UUID string `json:"uuid"`
}

// AgentInfo is local metadata included in registration.
type AgentInfo struct {
	UUID         string
	Hostname     string
	OS           string
	Architecture string
	Version      string
}

// Client maintains a WebSocket connection to the DockSight server.
type Client struct {
	serverURL      string
	info           AgentInfo
	heartbeatEvery time.Duration
	reconnectWait  time.Duration
}

// NewClient creates a reconnecting agent communication client.
func NewClient(serverURL string, info AgentInfo) *Client {
	return &Client{
		serverURL:      serverURL,
		info:           info,
		heartbeatEvery: 30 * time.Second,
		reconnectWait:  5 * time.Second,
	}
}

// Run connects, registers, heartbeats, and reconnects until ctx is cancelled.
func (c *Client) Run(ctx context.Context) {
	attempt := 0
	for {
		if ctx.Err() != nil {
			return
		}

		attempt++
		logger.Info("connecting to docksight server", "url", c.serverURL, "attempt", attempt)
		if err := c.session(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("agent session ended; reconnecting", "error", err.Error(), "retryIn", c.reconnectWait.String())
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(c.reconnectWait):
		}
	}
}

func (c *Client) session(ctx context.Context) error {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return fmt.Errorf("invalid server url: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}
	defer conn.Close()

	logger.Info("websocket connected", "url", c.serverURL)

	if err := c.sendRegister(conn); err != nil {
		return err
	}

	registered, err := c.waitRegistered(conn, 15*time.Second)
	if err != nil {
		return err
	}
	logger.Info("registration acknowledged",
		"id", registered.ID,
		"uuid", registered.UUID,
		"status", registered.Status,
		"message", registered.Message,
	)

	return c.heartbeatLoop(ctx, conn)
}

func (c *Client) sendRegister(conn *websocket.Conn) error {
	payload, err := json.Marshal(RegisterPayload{
		UUID:         c.info.UUID,
		Hostname:     c.info.Hostname,
		OS:           c.info.OS,
		Architecture: c.info.Architecture,
		Version:      c.info.Version,
	})
	if err != nil {
		return err
	}
	return c.write(conn, Envelope{Type: TypeRegister, Payload: payload})
}

func (c *Client) waitRegistered(conn *websocket.Conn, timeout time.Duration) (*RegisteredPayload, error) {
	deadline := time.Now().Add(timeout)
	_ = conn.SetReadDeadline(deadline)
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("wait for registration response: %w", err)
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		if env.Type != TypeRegistered {
			continue
		}

		var payload RegisteredPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return nil, fmt.Errorf("parse registered payload: %w", err)
		}
		if payload.Message != "" && payload.ID == "" && payload.UUID == "" {
			return nil, fmt.Errorf("registration failed: %s", payload.Message)
		}
		return &payload, nil
	}
}

func (c *Client) heartbeatLoop(ctx context.Context, conn *websocket.Conn) error {
	ticker := time.NewTicker(c.heartbeatEvery)
	defer ticker.Stop()

	errCh := make(chan error, 1)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				errCh <- err
				return
			}
		}
	}()

	// Send an immediate heartbeat after registration.
	if err := c.sendHeartbeat(conn); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
			)
			return ctx.Err()
		case err := <-errCh:
			return fmt.Errorf("connection closed: %w", err)
		case <-ticker.C:
			if err := c.sendHeartbeat(conn); err != nil {
				return err
			}
			logger.Debug("heartbeat sent", "uuid", c.info.UUID)
		}
	}
}

func (c *Client) sendHeartbeat(conn *websocket.Conn) error {
	payload, err := json.Marshal(HeartbeatPayload{UUID: c.info.UUID})
	if err != nil {
		return err
	}
	return c.write(conn, Envelope{Type: TypeHeartbeat, Payload: payload})
}

func (c *Client) write(conn *websocket.Conn, env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
