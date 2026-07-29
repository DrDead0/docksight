package release

import (
	"fmt"
	"strings"
)

// installerAssetPrefix is the filename prefix of the installer bundle
// published with every release, e.g. docksight-install-v0.0.2.tar.gz
const installerAssetPrefix = "docksight-install-"

// FindInstallerAsset returns the installer bundle attached to the given
// release, or an error if the release does not publish one.
func FindInstallerAsset(githubRelease *GithubRelease) (*Asset, error) {

	if githubRelease == nil {
		return nil, fmt.Errorf("no release provided")
	}

	for i := range githubRelease.Assets {

		asset := &githubRelease.Assets[i]

		if strings.HasPrefix(asset.Name, installerAssetPrefix) {
			return asset, nil
		}
	}

	return nil, fmt.Errorf(
		"installer asset %s* not found in release %s",
		installerAssetPrefix,
		githubRelease.TagName,
	)
}
