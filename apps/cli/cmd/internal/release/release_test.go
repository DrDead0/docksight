package release

import (
	"testing"
)

func TestFindInstallerAsset(t *testing.T) {

	githubRelease := &GithubRelease{
		TagName: "v0.0.2",
		Assets: []Asset{
			{Name: "checksums.txt", URL: "https://example.com/checksums.txt"},
			{
				Name: "docksight-install-v0.0.2.tar.gz",
				URL:  "https://example.com/docksight-install-v0.0.2.tar.gz",
			},
		},
	}

	asset, err := FindInstallerAsset(githubRelease)

	if err != nil {
		t.Fatal(err)
	}

	if asset.Name != "docksight-install-v0.0.2.tar.gz" {
		t.Fatalf("picked the wrong asset: %q", asset.Name)
	}

	if asset.URL != "https://example.com/docksight-install-v0.0.2.tar.gz" {
		t.Fatalf("asset url mismatch: %q", asset.URL)
	}
}

func TestFindInstallerAssetMissing(t *testing.T) {

	githubRelease := &GithubRelease{
		TagName: "v0.0.2",
		Assets: []Asset{
			{Name: "checksums.txt", URL: "https://example.com/checksums.txt"},
		},
	}

	if _, err := FindInstallerAsset(githubRelease); err == nil {
		t.Fatal("expected an error when no installer asset is published")
	}
}

func TestFindInstallerAssetNoAssets(t *testing.T) {

	if _, err := FindInstallerAsset(&GithubRelease{TagName: "v0.0.2"}); err == nil {
		t.Fatal("expected an error for a release with no assets")
	}
}

func TestFindInstallerAssetNilRelease(t *testing.T) {

	if _, err := FindInstallerAsset(nil); err == nil {
		t.Fatal("expected an error for a nil release")
	}
}
