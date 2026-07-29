package release

import (
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"io"
	"net/http"
	"os"
)


func Download(url string, destination string) error {

	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		ui.Error(
			"download failed with status: %s"+
			resp.Status,
		)
		return nil
	}


	file, err := os.Create(destination)

	if err != nil {
		return err
	}

	defer file.Close()


	_, err = io.Copy(file, resp.Body)

	if err != nil {
		return err
	}


	return nil
}