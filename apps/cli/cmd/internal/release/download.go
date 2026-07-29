package release

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Download fetches url and writes the response body to destination,
// creating the parent directory if needed.
func Download(url string, destination string) error {

	resp, err := httpClient.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	file, err := os.Create(destination)

	if err != nil {
		return err
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		// Do not leave a truncated bundle behind for the extractor to find.
		file.Close()
		os.Remove(destination)
		return err
	}

	return file.Close()
}
