package logs

import (
	"context"
	"io"
)

// Entry is one decoded container log line.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Stream    string `json:"stream"`
	Message   string `json:"message"`
}

// SubscribeOptions configures a log stream from the protocol payload.
type SubscribeOptions struct {
	RequestID   string
	ContainerID string
	Tail        int
	Follow      bool
}

// Chunk is a batched set of log entries ready to send as logs.chunk.
type Chunk struct {
	RequestID   string
	ContainerID string
	Entries     []Entry
}

// ChunkEmitter sends batched log chunks to the DockSight server.
type ChunkEmitter interface {
	EmitLogChunk(chunk Chunk) error
}

// Engine is the Docker log source used by the logs service.
type Engine interface {
	ContainerLogs(ctx context.Context, containerID string, tail string, follow bool) (io.ReadCloser, error)
}
