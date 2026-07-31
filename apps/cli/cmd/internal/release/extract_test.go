package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

type entry struct {
	name    string
	content string
	dir     bool
	mode    int64
	typeFlg byte
}

// tarball builds a gzipped archive in memory.
func tarball(t *testing.T, entries ...entry) string {
	t.Helper()

	var buffer bytes.Buffer

	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, item := range entries {

		header := &tar.Header{
			Name:     item.name,
			Mode:     item.mode,
			Size:     int64(len(item.content)),
			Typeflag: item.typeFlg,
		}

		if header.Mode == 0 {
			header.Mode = 0o644
		}

		if item.dir {
			header.Typeflag = tar.TypeDir
			header.Size = 0
		} else if header.Typeflag == 0 {
			header.Typeflag = tar.TypeReg
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}

		if header.Typeflag == tar.TypeReg {
			if _, err := tarWriter.Write([]byte(item.content)); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}

	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "archive.tar.gz")

	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestExtractTarGz(t *testing.T) {

	archive := tarball(t,
		entry{name: "bundle/", dir: true, mode: 0o755},
		entry{name: "bundle/VERSION", content: "v0.0.4"},
		entry{name: "bundle/.env.example", content: "JWT_SECRET=x"},
		entry{name: "bundle/nested/default.conf", content: "server {}"},
	)

	destination := filepath.Join(t.TempDir(), "extracted")

	if err := ExtractTarGz(archive, destination); err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string]string{
		"bundle/VERSION":             "v0.0.4",
		"bundle/.env.example":        "JWT_SECRET=x",
		"bundle/nested/default.conf": "server {}",
	} {

		content, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(name)))

		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if string(content) != want {
			t.Errorf("%s: got %q", name, string(content))
		}
	}
}

// Install runs as root, so an entry climbing out of the destination would
// write anywhere on the host.
func TestExtractRejectsPathTraversal(t *testing.T) {

	archive := tarball(t, entry{name: "../escaped.txt", content: "owned"})

	destination := filepath.Join(t.TempDir(), "extracted")

	err := ExtractTarGz(archive, destination)

	if err == nil {
		t.Fatal("a traversing entry was extracted")
	}

	escaped := filepath.Join(filepath.Dir(destination), "escaped.txt")

	if _, statErr := os.Stat(escaped); statErr == nil {
		t.Fatalf("entry escaped to %s", escaped)
	}
}

func TestExtractRejectsAbsolutePaths(t *testing.T) {

	archive := tarball(t, entry{name: "/etc/cron.d/docksight", content: "* * * * * root sh"})

	if err := ExtractTarGz(archive, filepath.Join(t.TempDir(), "extracted")); err == nil {
		t.Fatal("an absolute entry was extracted")
	}
}

// A symlink entry is the other way out of the destination directory. Bundles
// contain none, so skipping them is safe and keeps extraction total.
func TestExtractSkipsSymlinks(t *testing.T) {

	archive := tarball(t,
		entry{name: "bundle/VERSION", content: "v0.0.4"},
		entry{name: "bundle/link", typeFlg: tar.TypeSymlink},
	)

	destination := filepath.Join(t.TempDir(), "extracted")

	if err := ExtractTarGz(archive, destination); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(destination, "bundle", "VERSION")); err != nil {
		t.Fatalf("extraction stopped at the symlink: %v", err)
	}
}

func TestExtractRejectsNonArchive(t *testing.T) {

	plain := filepath.Join(t.TempDir(), "not-an-archive.tar.gz")

	if err := os.WriteFile(plain, []byte("<html>404</html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ExtractTarGz(plain, filepath.Join(t.TempDir(), "out")); err == nil {
		t.Fatal("expected an error for a non-gzip file")
	}
}
