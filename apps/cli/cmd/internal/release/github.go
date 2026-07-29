package release

import (
	"encoding/json"
	"errors"
	"net/http"
)

func Latest() (*GithubRelease, error) {

	resp, err := http.Get(
		"https://api.github.com/repos/rodriguecyber/docksight/releases/latest",
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get latest release: " + resp.Status)
	}

	var data GithubRelease

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return &data, nil
}
