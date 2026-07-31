package release

import (
	"fmt"
	"strings"
)

// Naming convention for published release assets. These four constants are
// the single place in the codebase that knows what a DockSight asset is
// called; everything else asks for an asset by kind.
//
//	platform bundle : docksight-platform-v0.0.4.tar.gz
//	CLI binary      : docksight-cli-v0.0.4-linux-amd64
//	agent binary    : docksight-agent-v0.0.4-linux-amd64
const (
	platformPrefix = "docksight-platform-"
	cliPrefix      = "docksight-cli-"
	agentPrefix    = "docksight-agent-"
	archiveSuffix  = ".tar.gz"

	// legacyPlatformPrefix is the name used by releases up to v0.0.3, kept so
	// a new CLI can still install an older platform bundle.
	legacyPlatformPrefix = "docksight-install-"
)

// Asset is a file published with a GitHub release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

// Selector decides which asset of a release is wanted. Callers name what they
// need ("the CLI binary for linux/arm64") instead of matching filenames.
type Selector interface {
	// Matches reports whether the asset is the one being looked for.
	Matches(Asset) bool

	// Describe names the wanted asset for error messages.
	Describe() string
}

type selectorFunc struct {
	match    func(Asset) bool
	describe string
}

func (s selectorFunc) Matches(asset Asset) bool { return s.match(asset) }
func (s selectorFunc) Describe() string         { return s.describe }

// PlatformBundle selects the compose application bundle: compose file,
// nginx config, environment template and VERSION. It never contains the CLI.
func PlatformBundle() Selector {

	return selectorFunc{
		match: func(asset Asset) bool {

			if !strings.HasSuffix(asset.Name, archiveSuffix) {
				return false
			}

			return strings.HasPrefix(asset.Name, platformPrefix) ||
				strings.HasPrefix(asset.Name, legacyPlatformPrefix)
		},
		describe: platformPrefix + "*" + archiveSuffix,
	}
}

// CLIBinary selects the standalone CLI executable for one platform. The
// bundle never contains the CLI, so this is the only way to obtain it.
func CLIBinary(target Target) Selector {

	return selectorFunc{
		match: func(asset Asset) bool {

			if !strings.HasPrefix(asset.Name, cliPrefix) {
				return false
			}

			// A CLI asset may be shipped raw or archived; either is fine as
			// long as it is built for the requested platform.
			return target.matches(asset.Name)
		},
		describe: fmt.Sprintf("%s* for %s", cliPrefix, target),
	}
}

// AgentBinary selects the agent executable for one platform. The agent runs
// on remote Docker hosts and is versioned with the rest of the release.
func AgentBinary(target Target) Selector {

	return selectorFunc{
		match: func(asset Asset) bool {

			if !strings.HasPrefix(asset.Name, agentPrefix) {
				return false
			}

			return target.matches(asset.Name)
		},
		describe: fmt.Sprintf("%s* for %s", agentPrefix, target),
	}
}

// AgentBinaryName builds the filename the release pipeline should publish for
// an agent binary.
func AgentBinaryName(version string, target Target) string {
	return agentPrefix + version + "-" + target.String() + target.Extension()
}

// PlatformBundleName builds the filename the release pipeline should publish
// for a platform bundle. Kept here so packaging and discovery cannot drift.
func PlatformBundleName(version string) string {
	return platformPrefix + version + archiveSuffix
}

// CLIBinaryName builds the filename the release pipeline should publish for a
// CLI binary.
func CLIBinaryName(version string, target Target) string {
	return cliPrefix + version + "-" + target.String() + target.Extension()
}

// IsArchive reports whether the asset needs extracting after download.
func (a Asset) IsArchive() bool {
	return strings.HasSuffix(a.Name, archiveSuffix)
}
