package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

// bundle builds a directory tree shaped like an extracted platform bundle.
func bundle(t *testing.T) string {
	t.Helper()

	source := t.TempDir()

	if err := os.MkdirAll(filepath.Join(source, "compose"), 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"VERSION":               "v0.0.1",
		".env.example":          "PORT=2002",
		"compose/docksight.yml": "services: {}",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return source
}

func TestCopyDir(t *testing.T) {

	source := bundle(t)
	destination := filepath.Join(t.TempDir(), "install")

	if err := CopyDir(source, destination); err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string]string{
		"VERSION":               "v0.0.1",
		".env.example":          "PORT=2002",
		"compose/docksight.yml": "services: {}",
	} {

		got, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(name)))

		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if string(got) != want {
			t.Fatalf("%s: got %q, want %q", name, string(got), want)
		}
	}
}

func TestCopyDirOverwritesExistingInstall(t *testing.T) {

	source := bundle(t)
	destination := t.TempDir()

	// A previous install left a file in place. Its content only has to differ
	// from the bundle's VERSION — if the two matched, a CopyDir that silently
	// did nothing would pass this test.
	if err := os.WriteFile(filepath.Join(destination, "VERSION"), []byte("previous-install"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CopyDir(source, destination); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(destination, "VERSION"))

	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "v0.0.1" {
		t.Fatalf("stale file not replaced: %q", string(got))
	}
}

func TestCopyDirMissingSource(t *testing.T) {

	missing := filepath.Join(t.TempDir(), "no-bundle-here")

	if err := CopyDir(missing, t.TempDir()); err == nil {
		t.Fatal("expected an error when the source does not exist")
	}
}
