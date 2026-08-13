package release

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const checksumsName = "checksums.txt"

// ParseChecksums reads the `sha256sum` format published next to release assets.
func ParseChecksums(r io.Reader) (map[string]string, error) {
	out := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// sha256sum: "<hex>  <filename>" or "<hex> *<filename>"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("malformed checksums line: %q", line)
		}
		sum := strings.ToLower(fields[0])
		if len(sum) != 64 {
			return nil, fmt.Errorf("malformed checksum %q", fields[0])
		}
		name := strings.TrimPrefix(fields[1], "*")
		out[name] = sum
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// FileSHA256 returns the hex SHA-256 of the file at path.
func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// Checksums downloads and parses checksums.txt from a release.
// A missing file returns (nil, nil) so callers can warn rather than fail
// on releases published before checksums existed.
func (c *Client) Checksums(ctx context.Context, rel *Release) (map[string]string, error) {
	if rel == nil {
		return nil, fmt.Errorf("no release provided")
	}

	var asset *Asset
	for i := range rel.Assets {
		if rel.Assets[i].Name == checksumsName {
			asset = &rel.Assets[i]
			break
		}
	}
	if asset == nil {
		return nil, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.http().Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download %s: %s", checksumsName, response.Status)
	}

	return ParseChecksums(response.Body)
}

// VerifyDownloaded checks the on-disk file against checksums.txt.
// Missing checksums.txt is not an error; the bool is false when nothing
// could be verified so the caller can warn.
func (c *Client) VerifyDownloaded(ctx context.Context, rel *Release, name, path string) (verified bool, err error) {
	sums, err := c.Checksums(ctx, rel)
	if err != nil {
		return false, err
	}
	if sums == nil {
		return false, nil
	}
	expected, ok := sums[name]
	if !ok {
		return false, fmt.Errorf("checksums.txt has no entry for %s", name)
	}
	actual, err := FileSHA256(path)
	if err != nil {
		return false, err
	}
	if actual != expected {
		return false, fmt.Errorf("checksum mismatch for %s: expected %s, got %s", name, expected, actual)
	}
	return true, nil
}
