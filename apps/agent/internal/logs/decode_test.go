package logs

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"
)

func TestParseTimestampedLine(t *testing.T) {
	entry := ParseTimestampedLine("2026-07-25T10:00:00.123456789Z Application started", "stdout")
	if entry.Timestamp != "2026-07-25T10:00:00.123456789Z" {
		t.Fatalf("timestamp=%q", entry.Timestamp)
	}
	if entry.Message != "Application started" {
		t.Fatalf("message=%q", entry.Message)
	}
	if entry.Stream != "stdout" {
		t.Fatalf("stream=%q", entry.Stream)
	}
}

func TestDecodeMultiplexedStdoutStderr(t *testing.T) {
	var buf bytes.Buffer
	writeFrame(&buf, 1, []byte("2026-07-25T10:00:00Z hello stdout\n"))
	writeFrame(&buf, 2, []byte("2026-07-25T10:00:01Z hello stderr\n"))

	var entries []Entry
	if err := DecodeLogStream(&buf, func(entry Entry) error {
		entries = append(entries, entry)
		return nil
	}); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("entries=%d", len(entries))
	}
	if entries[0].Stream != "stdout" || entries[0].Message != "hello stdout" {
		t.Fatalf("entry0=%+v", entries[0])
	}
	if entries[1].Stream != "stderr" || entries[1].Message != "hello stderr" {
		t.Fatalf("entry1=%+v", entries[1])
	}
}

func TestDecodePlainLines(t *testing.T) {
	input := "2026-07-25T10:00:00Z line-one\n2026-07-25T10:00:01Z line-two\n"
	var entries []Entry
	if err := DecodeLogStream(bytes.NewBufferString(input), func(entry Entry) error {
		entries = append(entries, entry)
		return nil
	}); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries=%d %+v", len(entries), entries)
	}
	if entries[0].Message != "line-one" || entries[1].Message != "line-two" {
		t.Fatalf("messages=%q %q", entries[0].Message, entries[1].Message)
	}
}

func writeFrame(w io.Writer, stream byte, payload []byte) {
	var header [8]byte
	header[0] = stream
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	_, _ = w.Write(header[:])
	_, _ = w.Write(payload)
}

func TestParseTimestampedLineFallback(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	entry := ParseTimestampedLine("no-timestamp message", "stderr")
	after := time.Now().UTC().Add(time.Second)
	if entry.Message != "no-timestamp message" {
		t.Fatalf("message=%q", entry.Message)
	}
	ts, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
	if err != nil {
		t.Fatalf("parse ts: %v", err)
	}
	if ts.Before(before) || ts.After(after) {
		t.Fatalf("unexpected fallback timestamp %s", entry.Timestamp)
	}
}
