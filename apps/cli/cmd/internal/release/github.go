package release

import (
	"encoding/json"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
	"net/http"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func Latest() (*Release, error) {

	resp, err := http.Get(
		"https://api.github.com/repos/rodriguecyber/docksight/releases/latest",
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.Error("failed to get latest release")
		// return nil
	}

	var data githubRelease

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return &Release{
		Version: data.TagName,
	}, nil
}
