package env

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// secretMarkers identify variables whose value must never be the one shipped
// in the example file. Everything else (database names, ports) is copied as
// written, because compose references those values by name.
var secretMarkers = []string{
	"PASSWORD",
	"PASSWD",
	"SECRET",
	"TOKEN",
	"KEY",
	"SALT",
	"CREDENTIAL",
}

// secretBytes is the entropy used per generated value. 48 raw bytes render as
// 64 base64 characters, matching the JWT secret shipped in the example.
const secretBytes = 48

// Generate writes destination from the example template, replacing every
// secret-like value with freshly generated entropy and preserving comments,
// blank lines and key order.
//
// It reports false when destination already exists: regenerating credentials
// over a running installation would lock the services out of data volumes
// created with the previous password.
func Generate(examplePath string, destinationPath string) (bool, error) {

	example, err := os.Open(examplePath)

	if err != nil {
		return false, err
	}

	defer example.Close()

	// O_EXCL makes the "already installed" check atomic — a separate
	// os.Stat would leave a window where two runs both decide to write.
	destination, err := os.OpenFile(
		destinationPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)

	if err != nil {
		if os.IsExist(err) {
			return false, nil
		}

		return false, err
	}

	writer := bufio.NewWriter(destination)
	scanner := bufio.NewScanner(example)

	for scanner.Scan() {

		line, err := renderLine(scanner.Text())

		if err != nil {
			destination.Close()
			os.Remove(destinationPath)
			return false, err
		}

		if _, err := writer.WriteString(line + "\n"); err != nil {
			destination.Close()
			os.Remove(destinationPath)
			return false, err
		}
	}

	if err := scanner.Err(); err != nil {
		destination.Close()
		os.Remove(destinationPath)
		return false, err
	}

	if err := writer.Flush(); err != nil {
		destination.Close()
		os.Remove(destinationPath)
		return false, err
	}

	if err := destination.Close(); err != nil {
		return false, err
	}

	return true, nil
}

// renderLine returns the line to write for one line of the example file.
func renderLine(line string) (string, error) {

	trimmed := strings.TrimSpace(line)

	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line, nil
	}

	key, _, found := strings.Cut(line, "=")

	if !found {
		return line, nil
	}

	if !isSecret(key) {
		return line, nil
	}

	value, err := secret()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(key) + "=" + value, nil
}

func isSecret(key string) bool {

	name := strings.ToUpper(strings.TrimSpace(key))

	for _, marker := range secretMarkers {
		if strings.Contains(name, marker) {
			return true
		}
	}

	return false
}

// secret returns a cryptographically random value. It is base64url encoded so
// the result carries no character that a shell, a compose file or a connection
// string would treat as syntax.
func secret() (string, error) {

	buffer := make([]byte, secretBytes)

	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
