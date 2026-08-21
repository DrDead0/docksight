package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultOwner   = "Open-Source-Kigali"
	defaultRepo    = "docksight"
	defaultBaseURL = "https://api.github.com"
)

// Client reads releases from a GitHub repository. The zero value is not
// usable; call NewClient.
type Client struct {
	Owner string
	Repo  string

	// BaseURL is the GitHub API root. Overridden in tests.
	BaseURL string

	HTTP *http.Client
}

// NewClient returns a client for the DockSight repository.
func NewClient() *Client {

	return &Client{
		Owner:   defaultOwner,
		Repo:    defaultRepo,
		BaseURL: defaultBaseURL,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Latest returns the most recent published release.
func (c *Client) Latest(ctx context.Context) (*Release, error) {

	return c.fetch(ctx, fmt.Sprintf(
		"%s/repos/%s/%s/releases/latest",
		c.BaseURL, c.Owner, c.Repo,
	))
}

// ByTag returns a specific release, so a user can pin or roll back a version.
func (c *Client) ByTag(ctx context.Context, tag string) (*Release, error) {

	return c.fetch(ctx, fmt.Sprintf(
		"%s/repos/%s/%s/releases/tags/%s",
		c.BaseURL, c.Owner, c.Repo, tag,
	))
}

func (c *Client) fetch(ctx context.Context, url string) (*Release, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/vnd.github+json")

	response, err := c.http().Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to fetch release: github returned %s",
			response.Status,
		)
	}

	var release Release

	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

// Download writes an asset to destination, creating the parent directory.
func (c *Client) Download(ctx context.Context, asset Asset, destination string) error {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)

	if err != nil {
		return err
	}

	response, err := c.http().Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"failed to download %s: %s",
			asset.Name,
			response.Status,
		)
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	file, err := os.Create(destination)

	if err != nil {
		return err
	}

	written, err := io.Copy(file, response.Body)
	if err != nil {
		// Never leave a truncated artifact for the extractor to find.
		file.Close()
		os.Remove(destination)
		return err
	}

	if err := file.Close(); err != nil {
		os.Remove(destination)
		return err
	}

	if asset.Size > 0 && written != asset.Size {
		os.Remove(destination)
		return fmt.Errorf(
			"truncated download of %s: got %d bytes, expected %d",
			asset.Name,
			written,
			asset.Size,
		)
	}

	return nil
}

func (c *Client) http() *http.Client {

	if c.HTTP != nil {
		return c.HTTP
	}

	return http.DefaultClient
}
