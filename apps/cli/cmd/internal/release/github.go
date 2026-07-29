package release

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// githubLatestReleaseURL is the GitHub API endpoint for the latest published
// release. It is a variable so tests can point it at a local server.
var githubLatestReleaseURL = "https://api.github.com/repos/rodriguecyber/docksight/releases/latest"

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// LatestGithubRelease fetches the latest published release from the GitHub API.
func LatestGithubRelease() (*GithubRelease, error) {

	req, err := http.NewRequest(http.MethodGet, githubLatestReleaseURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to fetch latest release: github returned %s",
			resp.Status,
		)
	}

	var githubRelease GithubRelease

	if err := json.NewDecoder(resp.Body).Decode(&githubRelease); err != nil {
		return nil, fmt.Errorf("failed to decode github release response: %w", err)
	}

	return &githubRelease, nil
}
