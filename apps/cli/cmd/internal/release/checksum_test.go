package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChecksums(t *testing.T) {
	input := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  docksight-cli-v0.1.0-linux-amd64\n"
	sums, err := ParseChecksums(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if sums["docksight-cli-v0.1.0-linux-amd64"] != strings.Repeat("a", 64) {
		t.Fatalf("got %#v", sums)
	}
}

func TestVerifyDownloadedMismatch(t *testing.T) {
	payload := []byte("not-the-published-bytes")
	sum := sha256.Sum256([]byte("the-published-bytes"))
	hexSum := hex.EncodeToString(sum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(hexSum + "  docksight-cli\n"))
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "docksight-cli")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}

	client := NewClient()
	rel := &Release{TagName: "v0.1.0", Assets: []Asset{{Name: "checksums.txt", URL: server.URL}}}
	_, err := client.VerifyDownloaded(context.Background(), rel, "docksight-cli", path)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected mismatch, got %v", err)
	}
}

func TestVerifyDownloadedMissingChecksumsIsSkip(t *testing.T) {
	client := NewClient()
	rel := &Release{TagName: "v0.1.0", Assets: []Asset{{Name: "docksight-cli", URL: "http://example.invalid"}}}
	verified, err := client.VerifyDownloaded(context.Background(), rel, "docksight-cli", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if verified {
		t.Fatal("missing checksums.txt must skip, not claim verification")
	}
}

func TestDownloadRejectsTruncatedPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("short"))
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "asset.bin")
	err := NewClient().Download(context.Background(), Asset{
		Name: "asset.bin",
		URL:  server.URL,
		Size: 100,
	}, destination)
	if err == nil || !strings.Contains(err.Error(), "truncated") {
		t.Fatalf("expected truncated download error, got %v", err)
	}
	if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
		t.Fatal("truncated download must not leave a file behind")
	}
}
