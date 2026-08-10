package envpath

import "testing"

// PATH has a length limit, and an installer that appends on every run walks
// towards it. Every case here is a way the same directory can be written.
func TestContains(t *testing.T) {

	const directory = `C:\Program Files\DockSight`

	cases := map[string]struct {
		path string
		want bool
	}{
		"exact":                   {`C:\Windows;C:\Program Files\DockSight`, true},
		"different case":          {`C:\Windows;c:\program files\docksight`, true},
		"trailing separator":      {`C:\Program Files\DockSight\`, true},
		"surrounded by spaces":    {`C:\Windows; C:\Program Files\DockSight `, true},
		"only element":            {`C:\Program Files\DockSight`, true},
		"empty":                   {``, false},
		"absent":                  {`C:\Windows;C:\Windows\System32`, false},
		"prefix of another entry": {`C:\Program Files\DockSightOther`, false},
		"empty elements":          {`C:\Windows;;C:\Program Files\DockSight`, true},
	}

	for name, testCase := range cases {

		t.Run(name, func(t *testing.T) {

			if got := Contains(testCase.path, directory); got != testCase.want {
				t.Fatalf("got %v, want %v for %q", got, testCase.want, testCase.path)
			}
		})
	}
}

func TestAppend(t *testing.T) {

	const directory = `C:\Program Files\DockSight`

	cases := map[string]struct {
		path        string
		want        string
		wantChanged bool
	}{
		"appends to an existing path": {
			path:        `C:\Windows`,
			want:        `C:\Windows;C:\Program Files\DockSight`,
			wantChanged: true,
		},
		"an empty path becomes the directory": {
			path:        ``,
			want:        directory,
			wantChanged: true,
		},
		"a trailing separator is not doubled": {
			path:        `C:\Windows;`,
			want:        `C:\Windows;C:\Program Files\DockSight`,
			wantChanged: true,
		},
		"already present is left alone": {
			path:        `C:\Windows;C:\Program Files\DockSight`,
			want:        `C:\Windows;C:\Program Files\DockSight`,
			wantChanged: false,
		},
	}

	for name, testCase := range cases {

		t.Run(name, func(t *testing.T) {

			got, changed := Append(testCase.path, directory)

			if got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}

			if changed != testCase.wantChanged {
				t.Errorf("changed = %v, want %v", changed, testCase.wantChanged)
			}
		})
	}
}

// A second install must be a no-op, or PATH grows without bound.
func TestAppendIsIdempotent(t *testing.T) {

	const directory = `C:\Program Files\DockSight`

	once, _ := Append(`C:\Windows`, directory)
	twice, changed := Append(once, directory)

	if changed {
		t.Error("a second append reported a change")
	}

	if twice != once {
		t.Fatalf("a second append altered the path: %q", twice)
	}
}

// A variable like %SystemRoot% must survive: PATH is REG_EXPAND_SZ and the
// entries are meant to stay symbolic.
func TestAppendPreservesUnexpandedEntries(t *testing.T) {

	got, _ := Append(`%SystemRoot%;%SystemRoot%\System32`, `C:\Program Files\DockSight`)

	want := `%SystemRoot%;%SystemRoot%\System32;C:\Program Files\DockSight`

	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
