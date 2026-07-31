package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Interactive reports whether stdin is a terminal. A prompt written to a pipe
// blocks forever, so callers must fall back to flags when this is false.
func Interactive() bool {

	info, err := os.Stdin.Stat()

	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

// Ask prints a question and reads one line of input. An empty answer returns
// fallback, which may be empty.
func Ask(question string, fallback string) (string, error) {

	if fallback != "" {
		fmt.Printf("? %s [%s]: ", question, fallback)
	} else {
		fmt.Printf("? %s: ", question)
	}

	reader := bufio.NewReader(os.Stdin)

	answer, err := reader.ReadString('\n')

	if err != nil && answer == "" {
		return "", err
	}

	answer = strings.TrimSpace(answer)

	if answer == "" {
		return fallback, nil
	}

	return answer, nil
}
