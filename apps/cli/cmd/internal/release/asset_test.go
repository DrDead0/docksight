package release

import (
	"testing"
)

func release(assets ...string) *Release {

	built := &Release{TagName: "v0.0.4"}

	for _, name := range assets {
		built.Assets = append(built.Assets, Asset{
			Name: name,
			URL:  "https://example.com/" + name,
		})
	}

	return built
}

func TestPlatformBundleSelection(t *testing.T) {

	rel := release(
		"checksums.txt",
		"docksight-cli-v0.0.4-linux-amd64",
		"docksight-platform-v0.0.4.tar.gz",
	)

	asset, err := rel.PlatformBundle()

	if err != nil {
		t.Fatal(err)
	}

	if asset.Name != "docksight-platform-v0.0.4.tar.gz" {
		t.Fatalf("picked %q", asset.Name)
	}
}

// Releases up to v0.0.3 named the bundle docksight-install-*.
func TestPlatformBundleAcceptsLegacyName(t *testing.T) {

	asset, err := release("docksight-install-v0.0.3.tar.gz").PlatformBundle()

	if err != nil {
		t.Fatal(err)
	}

	if asset.Name != "docksight-install-v0.0.3.tar.gz" {
		t.Fatalf("picked %q", asset.Name)
	}
}

// The CLI must never be served as the platform bundle, whatever it is called.
func TestPlatformBundleIgnoresCLI(t *testing.T) {

	if _, err := release("docksight-cli-v0.0.4-linux-amd64.tar.gz").PlatformBundle(); err == nil {
		t.Fatal("a CLI archive was accepted as the platform bundle")
	}
}

func TestCLIBinarySelectionPerTarget(t *testing.T) {

	rel := release(
		"docksight-platform-v0.0.4.tar.gz",
		"docksight-cli-v0.0.4-linux-amd64",
		"docksight-cli-v0.0.4-linux-arm64",
		"docksight-cli-v0.0.4-darwin-arm64",
		"docksight-cli-v0.0.4-windows-amd64.exe",
	)

	cases := map[Target]string{
		{OS: "linux", Arch: "amd64"}:   "docksight-cli-v0.0.4-linux-amd64",
		{OS: "linux", Arch: "arm64"}:   "docksight-cli-v0.0.4-linux-arm64",
		{OS: "darwin", Arch: "arm64"}:  "docksight-cli-v0.0.4-darwin-arm64",
		{OS: "windows", Arch: "amd64"}: "docksight-cli-v0.0.4-windows-amd64.exe",
	}

	for target, want := range cases {

		asset, err := rel.CLIBinary(target)

		if err != nil {
			t.Fatalf("%s: %v", target, err)
		}

		if asset.Name != want {
			t.Errorf("%s: picked %q, want %q", target, asset.Name, want)
		}
	}
}

// amd64 must never be served to an arm64 machine: the binary would not run.
func TestCLIBinaryMissingTarget(t *testing.T) {

	rel := release("docksight-cli-v0.0.4-linux-amd64")

	_, err := rel.CLIBinary(Target{OS: "linux", Arch: "arm64"})

	if err == nil {
		t.Fatal("an amd64 build was served for arm64")
	}
}

func TestFindErrorListsAvailableAssets(t *testing.T) {

	_, err := release("checksums.txt").PlatformBundle()

	if err == nil {
		t.Fatal("expected an error")
	}

	if got := err.Error(); !contains(got, "checksums.txt") || !contains(got, "v0.0.4") {
		t.Fatalf("error is not diagnosable: %s", got)
	}
}

func TestNamesRoundTripThroughSelectors(t *testing.T) {

	target := Target{OS: "linux", Arch: "arm64"}

	rel := release(
		PlatformBundleName("v0.1.0"),
		CLIBinaryName("v0.1.0", target),
		AgentBinaryName("v0.1.0", target),
	)

	if _, err := rel.PlatformBundle(); err != nil {
		t.Fatalf("published bundle name is not discoverable: %v", err)
	}

	if _, err := rel.CLIBinary(target); err != nil {
		t.Fatalf("published CLI name is not discoverable: %v", err)
	}

	if _, err := rel.AgentBinary(target); err != nil {
		t.Fatalf("published agent name is not discoverable: %v", err)
	}
}

// The agent is published for Windows as well as Linux. Discovery appends .exe
// through Target.Extension(), so a release that published the agent without
// the extension would be undiscoverable on the platform it was built for.
func TestAgentBinaryNameForWindows(t *testing.T) {

	target := Target{OS: "windows", Arch: "amd64"}

	name := AgentBinaryName("v0.1.0", target)

	if name != "docksight-agent-v0.1.0-windows-amd64.exe" {
		t.Fatalf("got %q", name)
	}

	if _, err := release(name).AgentBinary(target); err != nil {
		t.Fatalf("published Windows agent name is not discoverable: %v", err)
	}
}

// A Windows host must never be served a Linux agent: the binary would not run.
func TestAgentBinaryDoesNotCrossPlatforms(t *testing.T) {

	rel := release(
		"docksight-agent-v0.1.0-linux-amd64",
		"docksight-agent-v0.1.0-windows-amd64.exe",
	)

	asset, err := rel.AgentBinary(Target{OS: "windows", Arch: "amd64"})

	if err != nil {
		t.Fatal(err)
	}

	if asset.Name != "docksight-agent-v0.1.0-windows-amd64.exe" {
		t.Fatalf("picked %q for windows/amd64", asset.Name)
	}
}

func TestCLIBinaryNameForWindows(t *testing.T) {

	name := CLIBinaryName("v0.1.0", Target{OS: "windows", Arch: "amd64"})

	if name != "docksight-cli-v0.1.0-windows-amd64.exe" {
		t.Fatalf("got %q", name)
	}
}

func TestTargetTokenMatching(t *testing.T) {

	target := Target{OS: "linux", Arch: "arm64"}

	if target.matches("docksight-cli-v1-linux-amd64") {
		t.Error("arm64 matched an amd64 asset")
	}

	if !target.matches("docksight_cli_v1_linux_arm64") {
		t.Error("underscore separators should be accepted")
	}

	if target.matches("docksight-cli-v1-linuxarm64") {
		t.Error("tokens must be delimited")
	}
}

func TestIsArchive(t *testing.T) {

	if !(Asset{Name: "docksight-platform-v1.tar.gz"}).IsArchive() {
		t.Error("tarball not recognised as an archive")
	}

	if (Asset{Name: "docksight-cli-v1-linux-amd64"}).IsArchive() {
		t.Error("raw binary treated as an archive")
	}
}

func contains(haystack string, needle string) bool {

	for index := 0; index+len(needle) <= len(haystack); index++ {
		if haystack[index:index+len(needle)] == needle {
			return true
		}
	}

	return false
}
