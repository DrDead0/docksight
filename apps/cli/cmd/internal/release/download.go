package release

import (
	"io"
	"net/http"
	"os"
)


func Download(
	url string,
	destination string,
) error {


	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()


	file, err := os.Create(destination)

	if err != nil {
		return err
	}

	defer file.Close()


	_, err = io.Copy(
		file,
		resp.Body,
	)

	return err
}