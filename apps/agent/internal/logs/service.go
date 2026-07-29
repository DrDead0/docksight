package logs

import (
	"context"
	"fmt"
	"sync"

	"docksight-agent/internal/logger"
)

// Service manages concurrent container log streams keyed by requestId.
type Service struct {
	engine  Engine
	emitter ChunkEmitter

	mu      sync.Mutex
	streams map[string]*Stream
	rootCtx context.Context
	cancel  context.CancelFunc
}

// NewService creates a logs service. Call SetEmitter before Subscribe.
func NewService(engine Engine) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		engine:  engine,
		streams: make(map[string]*Stream),
		rootCtx: ctx,
		cancel:  cancel,
	}
}

// SetEmitter wires the communication sender used for logs.chunk messages.
func (s *Service) SetEmitter(emitter ChunkEmitter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emitter = emitter
}

// Subscribe starts (or replaces) a log stream for requestId.
func (s *Service) Subscribe(opts SubscribeOptions) error {
	s.mu.Lock()
	emitter := s.emitter
	engine := s.engine
	root := s.rootCtx
	existing := s.streams[opts.RequestID]
	if existing != nil {
		delete(s.streams, opts.RequestID)
	}
	s.mu.Unlock()

	if existing != nil {
		existing.close()
	}

	if engine == nil {
		return fmt.Errorf("subscribe: docker engine unavailable")
	}
	if emitter == nil {
		return fmt.Errorf("subscribe: chunk emitter is not configured")
	}

	stream, err := newStream(root, engine, opts, emitter)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if prev, ok := s.streams[opts.RequestID]; ok {
		delete(s.streams, opts.RequestID)
		s.mu.Unlock()
		prev.close()
		s.mu.Lock()
	}
	s.streams[opts.RequestID] = stream
	s.mu.Unlock()

	logger.Info("log stream subscribed",
		"requestId", opts.RequestID,
		"containerId", opts.ContainerID,
		"tail", opts.Tail,
		"follow", opts.Follow,
	)
	return nil
}

// Unsubscribe stops a single stream by requestId.
func (s *Service) Unsubscribe(requestID string) error {
	if requestID == "" {
		return fmt.Errorf("unsubscribe: requestId is required")
	}

	s.mu.Lock()
	stream, ok := s.streams[requestID]
	if ok {
		delete(s.streams, requestID)
	}
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("unsubscribe: stream %q not found", requestID)
	}

	stream.close()
	logger.Info("log stream unsubscribed", "requestId", requestID)
	return nil
}

// UnsubscribeAll stops every active stream (e.g. on WebSocket disconnect).
func (s *Service) UnsubscribeAll() {
	s.mu.Lock()
	streams := make([]*Stream, 0, len(s.streams))
	for id, stream := range s.streams {
		streams = append(streams, stream)
		delete(s.streams, id)
	}
	s.mu.Unlock()

	for _, stream := range streams {
		stream.close()
	}
}

// ActiveCount returns the number of managed streams (testing / diagnostics).
func (s *Service) ActiveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.streams)
}

// HasStream reports whether requestId is currently active.
func (s *Service) HasStream(requestID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.streams[requestID]
	return ok
}

// Close cancels the service root context and all streams.
func (s *Service) Close() {
	s.UnsubscribeAll()
	if s.cancel != nil {
		s.cancel()
	}
}
