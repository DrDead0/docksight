package release

import (
	"testing"
)

func TestLatestRelease(t *testing.T) {

	release, err := Latest()

	if err != nil {
		t.Fatal(err)
	}


	t.Log("Version:", release.TagName)


	for _, asset := range release.Assets {
		t.Log("Asset:", asset.Name)
	}
}