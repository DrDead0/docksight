package env

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const example = `# DockSight local development infrastructure
POSTGRES_DB=docksight
POSTGRES_USER=docksight
POSTGRES_PASSWORD=docksight-password
POSTGRES_PORT=5432
REDIS_PORT=6371
JWT_SECRET=UliOy9JA-GoqJ0jrIk68vqv4GUdQduBH-_2Uoq4MEVK4KTibiKd6yAcodRh0mdkV
`

// generate writes the example template into a temp dir and renders it.
func generate(t *testing.T) (string, map[string]string) {
	t.Helper()

	dir := t.TempDir()
	examplePath := filepath.Join(dir, ".env.example")

	if err := os.WriteFile(examplePath, []byte(example), 0o644); err != nil {
		t.Fatal(err)
	}

	destinationPath := filepath.Join(dir, ".env")

	created, err := Generate(examplePath, destinationPath)

	if err != nil {
		t.Fatal(err)
	}

	if !created {
		t.Fatal("expected the file to be created")
	}

	return destinationPath, parse(t, destinationPath)
}

func parse(t *testing.T, path string) map[string]string {
	t.Helper()

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	values := map[string]string{}

	for _, line := range strings.Split(string(content), "\n") {

		key, value, found := strings.Cut(line, "=")

		if found && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			values[key] = value
		}
	}

	return values
}

func TestGeneratePreservesNonSecrets(t *testing.T) {

	_, values := generate(t)

	// Compose references these by value — a random database name or a random
	// port would produce a stack that cannot start.
	for key, want := range map[string]string{
		"POSTGRES_DB":   "docksight",
		"POSTGRES_USER": "docksight",
		"POSTGRES_PORT": "5432",
		"REDIS_PORT":    "6371",
	} {
		if values[key] != want {
			t.Errorf("%s: got %q, want %q", key, values[key], want)
		}
	}
}

func TestGenerateReplacesSecrets(t *testing.T) {

	_, values := generate(t)

	for key, shipped := range map[string]string{
		"POSTGRES_PASSWORD": "docksight-password",
		"JWT_SECRET":        "UliOy9JA-GoqJ0jrIk68vqv4GUdQduBH-_2Uoq4MEVK4KTibiKd6yAcodRh0mdkV",
	} {

		if values[key] == shipped {
			t.Errorf("%s still holds the value from the example file", key)
		}

		if len(values[key]) < 40 {
			t.Errorf("%s is too short to be a generated secret: %q", key, values[key])
		}

		// The value must not need quoting in a compose file or a DSN.
		if strings.ContainsAny(values[key], "$\"'`\\ /+=") {
			t.Errorf("%s contains characters that need escaping: %q", key, values[key])
		}
	}
}

func TestGenerateIsUniquePerInstall(t *testing.T) {

	_, first := generate(t)
	_, second := generate(t)

	for _, key := range []string{"POSTGRES_PASSWORD", "JWT_SECRET"} {
		if first[key] == second[key] {
			t.Errorf("%s repeated across two installs", key)
		}
	}
}

func TestGeneratePreservesLayout(t *testing.T) {

	path, _ := generate(t)

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")

	if lines[0] != "# DockSight local development infrastructure" {
		t.Errorf("comment not preserved: %q", lines[0])
	}

	if !strings.HasPrefix(lines[1], "POSTGRES_DB=") {
		t.Errorf("key order not preserved: %q", lines[1])
	}
}

func TestGenerateKeepsExistingFile(t *testing.T) {

	dir := t.TempDir()
	examplePath := filepath.Join(dir, ".env.example")
	destinationPath := filepath.Join(dir, ".env")

	if err := os.WriteFile(examplePath, []byte(example), 0o644); err != nil {
		t.Fatal(err)
	}

	// An install that already ran: its password matches the data volume.
	if err := os.WriteFile(destinationPath, []byte("POSTGRES_PASSWORD=in-use\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	created, err := Generate(examplePath, destinationPath)

	if err != nil {
		t.Fatal(err)
	}

	if created {
		t.Fatal("expected the existing file to be reported as untouched")
	}

	content, err := os.ReadFile(destinationPath)

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "POSTGRES_PASSWORD=in-use\n" {
		t.Fatalf("existing credentials were overwritten: %q", string(content))
	}
}

func TestGenerateMissingExample(t *testing.T) {

	dir := t.TempDir()

	_, err := Generate(filepath.Join(dir, "absent"), filepath.Join(dir, ".env"))

	if err == nil {
		t.Fatal("expected an error when the example file is missing")
	}
}

func TestGenerateFileMode(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits are not meaningful on windows")
	}

	path, _ := generate(t)

	info, err := os.Stat(path)

	if err != nil {
		t.Fatal(err)
	}

	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("credentials file is %04o, want 0600", mode)
	}
}

func TestIsSecret(t *testing.T) {

	secrets := []string{
		"POSTGRES_PASSWORD", "JWT_SECRET", "API_TOKEN",
		"ENCRYPTION_KEY", "PASSWORD_SALT", "jwt_secret",
	}

	for _, key := range secrets {
		if !isSecret(key) {
			t.Errorf("%s should be treated as a secret", key)
		}
	}

	plain := []string{"POSTGRES_DB", "POSTGRES_PORT", "REDIS_PORT", "LOG_LEVEL"}

	for _, key := range plain {
		if isSecret(key) {
			t.Errorf("%s should be copied verbatim", key)
		}
	}
}
