package release

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const latestReleaseBody = `{
	"tag_name": "v0.0.3",
	"assets": [
		{"name": "checksums.txt", "browser_download_url": "https://example.com/checksums.txt"},
		{"name": "docksight-install-v0.0.3.tar.gz", "browser_download_url": "https://example.com/docksight-install-v0.0.3.tar.gz"}
	]
}`

// withReleaseURL points LatestGithubRelease at a test server for the duration
// of a single test.
func withReleaseURL(t *testing.T, url string) {
	t.Helper()

	original := githubLatestReleaseURL
	githubLatestReleaseURL = url

	t.Cleanup(func() {
		githubLatestReleaseURL = original
	})
}

func TestLatestGithubRelease(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(latestReleaseBody))
		},
	))

	defer server.Close()

	withReleaseURL(t, server.URL)

	githubRelease, err := LatestGithubRelease()

	if err != nil {
		t.Fatal(err)
	}

	if githubRelease.TagName != "v0.0.3" {
		t.Fatalf("expected tag v0.0.3, got %q", githubRelease.TagName)
	}

	if len(githubRelease.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(githubRelease.Assets))
	}

	if githubRelease.Assets[0].URL != "https://example.com/checksums.txt" {
		t.Fatalf("asset url not decoded: %q", githubRelease.Assets[0].URL)
	}
}

func TestLatestGithubReleaseNotFound(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		},
	))

	defer server.Close()

	withReleaseURL(t, server.URL)

	if _, err := LatestGithubRelease(); err == nil {
		t.Fatal("expected an error for a 404 response")
	}
}

func TestLatestGithubReleaseInvalidJSON(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		},
	))

	defer server.Close()

	withReleaseURL(t, server.URL)

	if _, err := LatestGithubRelease(); err == nil {
		t.Fatal("expected an error for a malformed response body")
	}
}

// TestLatestGithubReleaseLive hits the real GitHub API. Skipped under -short.
func TestLatestGithubReleaseLive(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping live GitHub API test")
	}

	githubRelease, err := LatestGithubRelease()

	if err != nil {
		t.Fatal(err)
	}

	t.Log("Version:", githubRelease.TagName)

	for _, asset := range githubRelease.Assets {
		t.Log("Asset:", asset.Name)
	}
}
