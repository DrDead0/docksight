package release

import (
	"runtime"
	"strings"
)

// Target identifies the platform a binary asset is built for. Release assets
// are matched against a Target rather than against a literal filename, so
// publishing darwin/arm64 builds later requires no code change here.
type Target struct {
	OS   string
	Arch string
}

// CurrentTarget is the platform this binary was compiled for.
func CurrentTarget() Target {
	return Target{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

// String renders the target as it appears in asset filenames, e.g.
// "linux-amd64".
func (t Target) String() string {
	return t.OS + "-" + t.Arch
}

// Extension is the suffix a binary carries on this target.
func (t Target) Extension() string {

	if t.OS == "windows" {
		return ".exe"
	}

	return ""
}

// matches reports whether an asset filename is built for this target. Both
// "-" and "_" are accepted as separators so the release pipeline is free to
// use either convention.
func (t Target) matches(name string) bool {

	lower := strings.ToLower(name)

	return containsToken(lower, t.OS) && containsToken(lower, t.Arch)
}

// containsToken looks for a token delimited by the separators used in release
// filenames, so "arm64" does not match inside an unrelated word and "amd64"
// never matches an "arm64" asset.
func containsToken(name string, token string) bool {

	if token == "" {
		return false
	}

	for index := 0; index+len(token) <= len(name); index++ {

		if name[index:index+len(token)] != token {
			continue
		}

		if index > 0 && !isSeparator(name[index-1]) {
			continue
		}

		end := index + len(token)

		if end < len(name) && !isSeparator(name[end]) {
			continue
		}

		return true
	}

	return false
}

func isSeparator(character byte) bool {

	switch character {
	case '-', '_', '.', '/':
		return true
	}

	return false
}
