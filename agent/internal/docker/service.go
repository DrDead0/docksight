package docker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Service provides read-only Docker Engine discovery helpers.
type Service struct {
	client *Client
}

// NewService creates a discovery service around a Docker client.
func NewService(client *Client) *Service {
	return &Service{client: client}
}

type engineInfoResponse struct {
	OperatingSystem string `json:"OperatingSystem"`
	Architecture    string `json:"Architecture"`
	ServerVersion   string `json:"ServerVersion"`
}

type engineVersionResponse struct {
	Version string `json:"Version"`
}

type containerJSON struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	Status  string   `json:"Status"`
	State   string   `json:"State"`
}

// GetDockerInfo returns Docker Engine metadata.
func (s *Service) GetDockerInfo(ctx context.Context) (*Info, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var info engineInfoResponse
	if err := s.client.getJSON(ctx, "/info", &info); err != nil {
		return nil, fmt.Errorf("docker info: %w", err)
	}

	var version engineVersionResponse
	if err := s.client.getJSON(ctx, "/version", &version); err != nil {
		return nil, fmt.Errorf("docker version: %w", err)
	}

	ver := version.Version
	if ver == "" {
		ver = info.ServerVersion
	}

	return &Info{
		Version:      ver,
		OS:           info.OperatingSystem,
		Architecture: info.Architecture,
	}, nil
}

// ListContainers returns a summary of containers (all states).
func (s *Service) ListContainers(ctx context.Context) ([]Container, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var items []containerJSON
	if err := s.client.getJSON(ctx, "/containers/json?all=true", &items); err != nil {
		return nil, fmt.Errorf("docker container list: %w", err)
	}

	result := make([]Container, 0, len(items))
	for _, item := range items {
		name := ""
		if len(item.Names) > 0 {
			name = strings.TrimPrefix(item.Names[0], "/")
		}
		result = append(result, Container{
			ID:     item.ID,
			Name:   name,
			Image:  item.Image,
			Status: item.Status,
			State:  item.State,
		})
	}
	return result, nil
}

// Ping verifies Docker Engine connectivity.
func (s *Service) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.client.getJSON(ctx, "/_ping", nil); err != nil {
		// Older engines may not expose /_ping the same way; fall back to /version.
		if err2 := s.client.getJSON(ctx, "/version", &engineVersionResponse{}); err2 != nil {
			return fmt.Errorf("docker ping: %w", err)
		}
	}
	return nil
}
