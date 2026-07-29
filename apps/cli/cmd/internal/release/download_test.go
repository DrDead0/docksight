package release

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownload(t *testing.T) {

	const payload = "docksight-installer-bundle"

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(payload))
		},
	))

	defer server.Close()

	// Nested dir so the parent-directory creation is exercised too.
	destination := filepath.Join(t.TempDir(), "nested", "docksight-install-v0.0.3.tar.gz")

	if err := Download(server.URL, destination); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(destination)

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != payload {
		t.Fatalf("downloaded content mismatch: %q", string(content))
	}
}

func TestDownloadBadStatus(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		},
	))

	defer server.Close()

	destination := filepath.Join(t.TempDir(), "missing.tar.gz")

	if err := Download(server.URL, destination); err == nil {
		t.Fatal("expected an error for a 404 response")
	}

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatal("no file should be created when the download fails")
	}
}

func TestDownloadUnreachableHost(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))

	url := server.URL
	server.Close()

	destination := filepath.Join(t.TempDir(), "unreachable.tar.gz")

	if err := Download(url, destination); err == nil {
		t.Fatal("expected an error when the host is unreachable")
	}
}
