package app

import (
	"io"
	"os"

	"docksight-agent/internal/logger"
	"docksight-agent/internal/service"
)

// openLogSink chooses where the agent's log goes.
//
// Run from a console — a developer, or a systemd unit whose stdout the journal
// captures — the answer is stdout, as it has always been. Run by the Windows
// Service Control Manager there is no console at all, so the log goes to a
// rotating file that survives the process and that "docksight agent install"
// can read back to confirm the agent connected.
//
// The returned close function is nil when there is nothing to close.
func openLogSink() (io.Writer, func(), error) {

	path := logger.ServiceLogPath()

	if !service.IsService() || path == "" {
		return os.Stdout, nil, nil
	}

	file, err := logger.NewRotatingFile(path)

	if err != nil {
		return nil, nil, err
	}

	return file, func() { _ = file.Close() }, nil
}
