package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const releaseBody = `{
	"tag_name": "v0.0.4",
	"assets": [
		{"name": "docksight-platform-v0.0.4.tar.gz", "browser_download_url": "https://example.com/platform", "size": 3382},
		{"name": "docksight-cli-v0.0.4-linux-amd64", "browser_download_url": "https://example.com/cli", "size": 9000000}
	]
}`

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)

	t.Cleanup(server.Close)

	client := NewClient()
	client.BaseURL = server.URL

	return client
}

func TestLatest(t *testing.T) {

	var requested string

	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		w.Write([]byte(releaseBody))
	})

	latest, err := client.Latest(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if requested != "/repos/Open-Source-Kigali/docksight/releases/latest" {
		t.Fatalf("requested %q", requested)
	}

	if latest.Version() != "v0.0.4" {
		t.Fatalf("got version %q", latest.Version())
	}

	if len(latest.Assets) != 2 {
		t.Fatalf("got %d assets", len(latest.Assets))
	}
}

func TestByTag(t *testing.T) {

	var requested string

	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		w.Write([]byte(releaseBody))
	})

	if _, err := client.ByTag(context.Background(), "v0.0.3"); err != nil {
		t.Fatal(err)
	}

	if requested != "/deliberately-broken-ci-check" {
		t.Fatalf("requested %q", requested)
	}
}

func TestLatestNotFound(t *testing.T) {

	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	if _, err := client.Latest(context.Background()); err == nil {
		t.Fatal("expected an error for a 404")
	}
}

func TestLatestMalformedBody(t *testing.T) {

	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})

	if _, err := client.Latest(context.Background()); err == nil {
		t.Fatal("expected an error for an unparseable body")
	}
}

func TestDownload(t *testing.T) {

	const payload = "platform-bundle-bytes"

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(payload))
		},
	))

	defer server.Close()

	client := NewClient()

	// Nested path so parent creation is exercised.
	destination := filepath.Join(t.TempDir(), "nested", "bundle.tar.gz")

	asset := Asset{Name: "bundle.tar.gz", URL: server.URL}

	if err := client.Download(context.Background(), asset, destination); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(destination)

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != payload {
		t.Fatalf("got %q", string(content))
	}
}

func TestDownloadBadStatus(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "gone", http.StatusGone)
		},
	))

	defer server.Close()

	destination := filepath.Join(t.TempDir(), "bundle.tar.gz")

	err := NewClient().Download(
		context.Background(),
		Asset{Name: "bundle.tar.gz", URL: server.URL},
		destination,
	)

	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "bundle.tar.gz") {
		t.Fatalf("error does not name the asset: %v", err)
	}

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatal("a failed download must not leave a file behind")
	}
}

func TestDownloadCancelled(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))

	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewClient().Download(
		ctx,
		Asset{Name: "bundle.tar.gz", URL: server.URL},
		filepath.Join(t.TempDir(), "bundle.tar.gz"),
	)

	if err == nil {
		t.Fatal("expected a cancelled download to fail")
	}
}
