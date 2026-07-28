package logs

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeEngine struct {
	mu      sync.Mutex
	readers map[string]*fakeReader
}

type fakeReader struct {
	io.Reader
	closed bool
	closeCh chan struct{}
}

func (r *fakeReader) Close() error {
	if !r.closed {
		r.closed = true
		close(r.closeCh)
	}
	return nil
}

func newFakeEngine() *fakeEngine {
	return &fakeEngine{readers: make(map[string]*fakeReader)}
}

func (e *fakeEngine) ContainerLogs(
	ctx context.Context,
	containerID string,
	tail string,
	follow bool,
) (io.ReadCloser, error) {
	_ = ctx
	_ = tail
	_ = follow
	body := "2026-07-25T10:00:00Z boot " + containerID + "\n"
	reader := &fakeReader{
		Reader:  strings.NewReader(body),
		closeCh: make(chan struct{}),
	}
	e.mu.Lock()
	e.readers[containerID] = reader
	e.mu.Unlock()
	return reader, nil
}

type recordingEmitter struct {
	mu     sync.Mutex
	chunks []Chunk
}

func (e *recordingEmitter) EmitLogChunk(chunk Chunk) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	copied := Chunk{
		RequestID:   chunk.RequestID,
		ContainerID: chunk.ContainerID,
		Entries:     append([]Entry(nil), chunk.Entries...),
	}
	e.chunks = append(e.chunks, copied)
	return nil
}

func (e *recordingEmitter) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.chunks)
}

func TestSubscribeCreatesStream(t *testing.T) {
	engine := newFakeEngine()
	emitter := &recordingEmitter{}
	svc := NewService(engine)
	svc.SetEmitter(emitter)
	defer svc.Close()

	err := svc.Subscribe(SubscribeOptions{
		RequestID:   "req-1",
		ContainerID: "backend",
		Tail:        50,
		Follow:      false,
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if !svc.HasStream("req-1") {
		t.Fatal("expected active stream")
	}
	if svc.ActiveCount() != 1 {
		t.Fatalf("active=%d", svc.ActiveCount())
	}

	waitFor(t, 2*time.Second, func() bool { return emitter.count() >= 1 })
}

func TestUnsubscribeCancelsOnlyTargetStream(t *testing.T) {
	engine := newFakeEngine()
	emitter := &recordingEmitter{}
	svc := NewService(engine)
	svc.SetEmitter(emitter)
	defer svc.Close()

	if err := svc.Subscribe(SubscribeOptions{RequestID: "req-1", ContainerID: "backend", Follow: true}); err != nil {
		t.Fatalf("subscribe 1: %v", err)
	}
	if err := svc.Subscribe(SubscribeOptions{RequestID: "req-2", ContainerID: "postgres", Follow: true}); err != nil {
		t.Fatalf("subscribe 2: %v", err)
	}
	if svc.ActiveCount() != 2 {
		t.Fatalf("active=%d", svc.ActiveCount())
	}

	if err := svc.Unsubscribe("req-1"); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if svc.HasStream("req-1") {
		t.Fatal("req-1 should be gone")
	}
	if !svc.HasStream("req-2") {
		t.Fatal("req-2 should remain")
	}
	if svc.ActiveCount() != 1 {
		t.Fatalf("active=%d", svc.ActiveCount())
	}
}

func TestMultipleStreamsIsolation(t *testing.T) {
	engine := newFakeEngine()
	emitter := &recordingEmitter{}
	svc := NewService(engine)
	svc.SetEmitter(emitter)
	defer svc.Close()

	if err := svc.Subscribe(SubscribeOptions{RequestID: "req-a", ContainerID: "backend", Follow: false}); err != nil {
		t.Fatalf("subscribe a: %v", err)
	}
	if err := svc.Subscribe(SubscribeOptions{RequestID: "req-b", ContainerID: "postgres", Follow: false}); err != nil {
		t.Fatalf("subscribe b: %v", err)
	}

	waitFor(t, 2*time.Second, func() bool { return emitter.count() >= 2 })

	emitter.mu.Lock()
	defer emitter.mu.Unlock()
	seen := map[string]bool{}
	for _, chunk := range emitter.chunks {
		seen[chunk.RequestID] = true
		if chunk.RequestID == "req-a" && chunk.ContainerID != "backend" {
			t.Fatalf("req-a mapped to %s", chunk.ContainerID)
		}
		if chunk.RequestID == "req-b" && chunk.ContainerID != "postgres" {
			t.Fatalf("req-b mapped to %s", chunk.ContainerID)
		}
	}
	if !seen["req-a"] || !seen["req-b"] {
		t.Fatalf("seen=%v", seen)
	}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
