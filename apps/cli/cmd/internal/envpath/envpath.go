// Package envpath puts the CLI's install directory on the system search path.
//
// It exists for Windows. On Linux the CLI installs into /usr/local/bin, which
// every shell already searches, so there is nothing to do and the non-Windows
// implementation says so by doing nothing. On Windows there is no such
// directory: an executable dropped in Program Files is not on PATH, and a
// user who installs and then types "docksight" in a new terminal is told the
// command does not exist.
package envpath

import "strings"

// SamePathElement reports whether two PATH entries name the same directory.
//
// Comparison is case-insensitive and ignores a trailing separator, because
// Windows paths are case-insensitive and "C:\Program Files\DockSight" and
// "C:\Program Files\DockSight\" are the same directory written twice. Getting
// this wrong appends a duplicate on every install, and PATH has a length
// limit that duplicates eventually reach.
func SamePathElement(left string, right string) bool {

	return normalizeElement(left) == normalizeElement(right)
}

func normalizeElement(element string) string {

	trimmed := strings.TrimSpace(element)
	trimmed = strings.TrimRight(trimmed, `\/`)

	return strings.ToLower(trimmed)
}

// Contains reports whether a PATH value already lists directory.
func Contains(path string, directory string) bool {

	for _, element := range strings.Split(path, ";") {

		if element == "" {
			continue
		}

		if SamePathElement(element, directory) {
			return true
		}
	}

	return false
}

// Append adds directory to a PATH value, returning the new value and whether
// it changed.
func Append(path string, directory string) (string, bool) {

	if Contains(path, directory) {
		return path, false
	}

	trimmed := strings.TrimRight(path, "; ")

	if trimmed == "" {
		return directory, true
	}

	return trimmed + ";" + directory, true
}
