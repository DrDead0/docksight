package filesystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// secretFile writes a file holding something that must not leak, using the
// permissive mode a careless caller might.
func secretFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".env")

	if err := os.WriteFile(path, []byte("POSTGRES_PASSWORD=hunter2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestProtectSecretUnix(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits are not meaningful on windows")
	}

	path := secretFile(t)

	if err := ProtectSecret(path); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)

	if err != nil {
		t.Fatal(err)
	}

	if permissions := info.Mode().Perm(); permissions&0o077 != 0 {
		t.Fatalf("group or world can still reach the file: %04o", permissions)
	}
}

// The Windows half is the reason this package exists: os.WriteFile's mode is
// ignored there, so a .env under ProgramData inherits a DACL that grants
// Users read — and POSTGRES_PASSWORD and JWT_SECRET become readable by every
// account on the machine.
//
// icacls is used to read the result back rather than the Win32 API, because
// asserting with the same calls that wrote it would pass even if both were
// wrong about what they mean.
func TestProtectSecretWindows(t *testing.T) {

	if runtime.GOOS != "windows" {
		t.Skip("windows ACLs")
	}

	path := secretFile(t)

	// The premise, logged rather than asserted so the test still means
	// something on a host whose defaults are already strict. On an ordinary
	// machine this prints the inherited entries that make the file readable —
	// and writable — by principals that have no business with a database
	// password.
	t.Logf("before:\n%s", icacls(t, path))

	if err := ProtectSecret(path); err != nil {
		t.Fatal(err)
	}

	after := icacls(t, path)

	// Asserting the whole trustee list rather than hunting for known-bad
	// names. Trustee names are localized and unresolvable accounts appear as
	// raw SIDs, so a deny-list would silently pass on the machines it most
	// needs to catch. Exactly two identities may appear.
	allowed := map[string]bool{
		`BUILTIN\Administrators`: true,
		`NT AUTHORITY\SYSTEM`:    true,
	}

	found := trustees(after)

	if len(found) == 0 {
		t.Fatalf("no access entries were parsed from:\n%s", after)
	}

	for _, trustee := range found {

		if !allowed[trustee] {
			t.Errorf("%q can still reach the credentials:\n%s", trustee, after)
		}
	}

	for required := range allowed {

		if !strings.Contains(after, required) {
			t.Errorf("%s lost access to the credentials:\n%s", required, after)
		}
	}

	// Inheritance must be off, or the two entries above would have been added
	// to the inherited ones rather than replacing them.
	if strings.Contains(after, "(I)") {
		t.Errorf("the file still inherits its parent's access list:\n%s", after)
	}
}

// trustees returns the identity named by each access entry in icacls output.
// Every entry is rendered as "TRUSTEE:(flags)", one per line.
func trustees(acl string) []string {

	names := make([]string, 0)

	for _, line := range strings.Split(acl, "\n") {

		name, _, found := strings.Cut(strings.TrimSpace(line), ":(")

		if !found || name == "" {
			continue
		}

		names = append(names, name)
	}

	return names
}

// icacls returns the access entries for a file, with the path itself removed.
//
// icacls echoes the path before the first entry, and a temporary directory
// under C:\Users contains the very word this test searches the entries for.
// Leaving it in makes every run report a false positive.
func icacls(t *testing.T, path string) string {
	t.Helper()

	output, err := exec.Command("icacls", path).CombinedOutput()

	if err != nil {
		t.Fatalf("icacls failed: %v\n%s", err, output)
	}

	return strings.ReplaceAll(string(output), path, "")
}
