// Package buildinfo carries values stamped into the binary at build time.
package buildinfo

var (
	// Version is the CLI's own version, independent of the platform bundle
	// version recorded in the install state.
	Version = "dev"

	// Commit is the source control revision used to build the binary.
	Commit = "dev"

	// BuildDate is the UTC timestamp of the build.
	BuildDate = "unknown"
)
