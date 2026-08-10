package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// maxLogBytes is the size a log file reaches before it is rotated.
	maxLogBytes int64 = 10 << 20 // 10 MiB

	// keptLogFiles is how many rotated files are retained alongside the
	// current one, bounding disk use at roughly (keptLogFiles+1) * maxLogBytes.
	keptLogFiles = 3
)

// RotatingFile is an io.WriteCloser that caps how much disk the agent's log
// can consume.
//
// A service runs unattended for months, so an unrotated file is a slow disk
// exhaustion bug on the host DockSight is supposed to be monitoring. Rotation
// is by size rather than by date because the agent's output rate depends on
// how much Docker activity it sees, not on the calendar.
type RotatingFile struct {
	mu sync.Mutex

	path     string
	maxBytes int64
	keep     int

	file *os.File
	size int64
}

// NewRotatingFile opens path for appending, creating its directory if needed.
func NewRotatingFile(path string) (*RotatingFile, error) {

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	rotating := &RotatingFile{
		path:     path,
		maxBytes: maxLogBytes,
		keep:     keptLogFiles,
	}

	if err := rotating.open(); err != nil {
		return nil, err
	}

	return rotating, nil
}

func (r *RotatingFile) open() error {

	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)

	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	// Append continues an existing file across restarts, so the size the
	// rotation check starts from is the size already on disk.
	info, err := file.Stat()

	if err != nil {
		file.Close()
		return fmt.Errorf("stat log file: %w", err)
	}

	r.file = file
	r.size = info.Size()

	return nil
}

// Write appends to the current file, rotating first when it is full.
func (r *RotatingFile) Write(p []byte) (int, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.size+int64(len(p)) > r.maxBytes {

		// A failed rotation must not lose the line being written: fall through
		// and keep appending to the file that is already open.
		if err := r.rotate(); err != nil {
			written, writeErr := r.file.Write(p)
			r.size += int64(written)

			if writeErr != nil {
				return written, writeErr
			}

			return written, nil
		}
	}

	written, err := r.file.Write(p)
	r.size += int64(written)

	return written, err
}

// rotate shifts agent.log.N to agent.log.N+1, moves the current file to .1 and
// opens a fresh one. The oldest file is discarded.
func (r *RotatingFile) rotate() error {

	if err := r.file.Close(); err != nil {
		return err
	}

	// The oldest file falls off the end.
	_ = os.Remove(fmt.Sprintf("%s.%d", r.path, r.keep))

	// Shift the rest up, highest first so no rename overwrites a file that
	// still has to be moved.
	for index := r.keep - 1; index >= 1; index-- {
		_ = os.Rename(
			fmt.Sprintf("%s.%d", r.path, index),
			fmt.Sprintf("%s.%d", r.path, index+1),
		)
	}

	if err := os.Rename(r.path, r.path+".1"); err != nil && !os.IsNotExist(err) {
		// Reopen so the writer stays usable even though rotation failed.
		_ = r.open()
		return err
	}

	return r.open()
}

// Close releases the underlying file.
func (r *RotatingFile) Close() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return nil
	}

	return r.file.Close()
}
