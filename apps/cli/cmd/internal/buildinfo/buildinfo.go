// Package buildinfo carries values stamped into the binary at build time.
package buildinfo

// Version is the CLI's own version, independent of the platform bundle
// version recorded in the install state. Set it at build time with:
//
//	go build -ldflags "-X github.com/rodriguecyber/docksight/apps/cli/cmd/internal/buildinfo.Version=v0.0.4"
var Version = "dev"
