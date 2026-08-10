package version

import "strings"

// Build-time values can be overridden with -ldflags, for example:
//
//	go build -ldflags "-X docksight-agent/internal/version.Version=0.1.0 -X docksight-agent/internal/version.Commit=abc123 -X docksight-agent/internal/version.BuildDate=2026-07-23"
var (
	// Version is the semantic version of the agent.
	Version = "0.1.0"

	// Commit is the source control revision used to build the binary.
	Commit = "dev"

	// BuildDate is the UTC timestamp of the build.
	BuildDate = "unknown"
)

// Info returns a snapshot of version metadata.
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildDate": BuildDate,
	}
}

// String returns a short human-readable version line.
//
// The release script stamps Version straight from the git tag, which carries
// a leading "v" ("v0.0.1"), while the unstamped default has none ("0.1.0").
// Trimming before prefixing gives each exactly one "v" — without it, a
// released build reported "vv0.0.1".
func String() string {
	return "v" + strings.TrimPrefix(Version, "v")
}
