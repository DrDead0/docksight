package version

import "testing"

// The release script stamps Version from the git tag, which carries a leading
// "v". Prefixing unconditionally produced "vv0.0.14" on every published build.
func TestStringNormalisesTagPrefix(t *testing.T) {
	t.Parallel()

	original := Version
	t.Cleanup(func() { Version = original })

	cases := map[string]string{
		"v0.0.14": "v0.0.14",
		"0.0.14":  "v0.0.14",
		"0.1.0":   "v0.1.0",
	}

	for stamped, want := range cases {
		Version = stamped

		if got := String(); got != want {
			t.Errorf("Version = %q: String() = %q, want %q", stamped, got, want)
		}
	}
}
