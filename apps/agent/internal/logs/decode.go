package logs

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	streamStdout = 1
	streamStderr = 2
)

// DecodeLogStream reads a Docker Engine log stream and invokes onEntry for each line.
// Supports multiplexed (non-TTY) frames and plain line-based (TTY) streams.
func DecodeLogStream(r io.Reader, onEntry func(Entry) error) error {
	br := bufio.NewReader(r)
	header, err := br.Peek(8)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("peek log header: %w", err)
	}

	// Multiplexed streams start with stream type 0/1/2 and a big-endian size.
	if isLikelyMultiplexed(header) {
		return decodeMultiplexed(br, onEntry)
	}
	return decodePlainLines(br, "stdout", onEntry)
}

func isLikelyMultiplexed(header []byte) bool {
	if len(header) < 8 {
		return false
	}
	streamType := header[0]
	if streamType > 2 {
		return false
	}
	// Bytes 1-3 are reserved zeroes in the Docker multiplex header.
	if header[1] != 0 || header[2] != 0 || header[3] != 0 {
		return false
	}
	return true
}

func decodeMultiplexed(br *bufio.Reader, onEntry func(Entry) error) error {
	header := make([]byte, 8)
	for {
		if _, err := io.ReadFull(br, header); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return fmt.Errorf("read multiplex header: %w", err)
		}

		streamName := streamName(header[0])
		size := binary.BigEndian.Uint32(header[4:8])
		if size == 0 {
			continue
		}

		payload := make([]byte, size)
		if _, err := io.ReadFull(br, payload); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return fmt.Errorf("read multiplex payload: %w", err)
		}

		if err := emitLines(string(payload), streamName, onEntry); err != nil {
			return err
		}
	}
}

func decodePlainLines(br *bufio.Reader, stream string, onEntry func(Entry) error) error {
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			if err := emitLines(line, stream, onEntry); err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read plain log line: %w", err)
		}
	}
}

func emitLines(payload string, stream string, onEntry func(Entry) error) error {
	payload = strings.ReplaceAll(payload, "\r\n", "\n")
	parts := strings.Split(payload, "\n")
	for i, part := range parts {
		// Keep empty trailing segment only when payload did not end with newline?
		// Drop empty fragments from trailing newlines.
		if part == "" && i == len(parts)-1 {
			continue
		}
		entry := ParseTimestampedLine(part, stream)
		if err := onEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

// ParseTimestampedLine splits a Docker `--timestamps` log line into timestamp + message.
func ParseTimestampedLine(line string, stream string) Entry {
	line = strings.TrimRight(line, "\r\n")
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	message := line

	space := strings.IndexByte(line, ' ')
	if space > 0 {
		candidate := line[:space]
		if _, err := time.Parse(time.RFC3339Nano, candidate); err == nil {
			timestamp = candidate
			message = line[space+1:]
		} else if t, err := time.Parse(time.RFC3339, candidate); err == nil {
			timestamp = t.UTC().Format(time.RFC3339Nano)
			message = line[space+1:]
		}
	}

	return Entry{
		Timestamp: timestamp,
		Stream:    stream,
		Message:   message,
	}
}

func streamName(code byte) string {
	switch code {
	case streamStdout:
		return "stdout"
	case streamStderr:
		return "stderr"
	default:
		return "stdout"
	}
}
