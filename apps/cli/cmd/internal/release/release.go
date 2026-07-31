package release

import (
	"fmt"
	"strings"
)

// Release is a published GitHub release and the assets attached to it.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Version is the release tag, e.g. "v0.0.4".
func (r *Release) Version() string {
	return r.TagName
}

// Find returns the single asset matching the selector.
func (r *Release) Find(selector Selector) (*Asset, error) {

	if r == nil {
		return nil, fmt.Errorf("no release provided")
	}

	for index := range r.Assets {

		if selector.Matches(r.Assets[index]) {
			return &r.Assets[index], nil
		}
	}

	return nil, fmt.Errorf(
		"release %s publishes no asset matching %s (has: %s)",
		r.TagName,
		selector.Describe(),
		r.assetNames(),
	)
}

// PlatformBundle returns the compose application bundle for this release.
func (r *Release) PlatformBundle() (*Asset, error) {
	return r.Find(PlatformBundle())
}

// CLIBinary returns the CLI executable built for the given target.
func (r *Release) CLIBinary(target Target) (*Asset, error) {
	return r.Find(CLIBinary(target))
}

// AgentBinary returns the agent executable built for the given target.
func (r *Release) AgentBinary(target Target) (*Asset, error) {
	return r.Find(AgentBinary(target))
}

func (r *Release) assetNames() string {

	if len(r.Assets) == 0 {
		return "no assets"
	}

	names := make([]string, 0, len(r.Assets))

	for _, asset := range r.Assets {
		names = append(names, asset.Name)
	}

	return strings.Join(names, ", ")
}
