package install

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Layout describes where the agent lives on a host.
type Layout struct {
	// BinaryPath is the installed agent executable.
	BinaryPath string

	// ConfigDir holds config.yaml and identity.json.
	ConfigDir string

	// ServiceName is the systemd unit, or the Service Control Manager service.
	ServiceName string

	// DockerSocket is the Engine socket, or named pipe, the agent talks to.
	DockerSocket string

	// LogLevel is the agent's log verbosity.
	LogLevel string

	// LogPath is where the agent writes its log. It is empty on Linux, where
	// the agent writes to stdout and systemd captures it into the journal.
	LogPath string

	// Windows is whether this layout describes a Windows host.
	//
	// It is the one flag the rest of the package branches on, and it says
	// something about the host the agent will run on rather than the host the
	// CLI was built for. Two things follow from it: how paths are joined into
	// the generated configuration, and whether a service definition is a
	// systemd unit or a command line for the Service Control Manager.
	Windows bool
}

// DefaultLayout is the standard agent installation for this host.
func DefaultLayout() Layout {

	if runtime.GOOS == "windows" {
		return WindowsLayout()
	}

	return LinuxLayout()
}

// LinuxLayout is the standard installation on a Linux host.
func LinuxLayout() Layout {

	return Layout{
		BinaryPath:   "/usr/local/bin/docksight-agent",
		ConfigDir:    "/etc/docksight-agent",
		ServiceName:  "docksight-agent.service",
		DockerSocket: "/var/run/docker.sock",
		LogLevel:     "info",
	}
}

// WindowsLayout is the standard installation on a Windows host.
//
// Program Files holds the executable and ProgramData holds the machine-wide
// state, which is the split Windows expects: the first is writable only by an
// installer, the second is where a service keeps what it must survive on.
// Both are read from the environment rather than hardcoded, because a machine
// may have them elsewhere and a localized Windows certainly does.
//
// The service name carries no ".service" suffix. That is systemd's filename
// convention, and the name here has to match the one the agent registers
// itself under in apps/agent/internal/service — the SCM matches on it, and a
// mismatch produces a service that starts and is immediately killed for
// failing to report in.
func WindowsLayout() Layout {

	return Layout{
		BinaryPath:   windowsJoin(programFiles(), "DockSight", "docksight-agent.exe"),
		ConfigDir:    windowsJoin(programData(), "DockSight", "agent"),
		ServiceName:  "docksight-agent",
		DockerSocket: `\\.\pipe\docker_engine`,
		LogLevel:     "info",
		LogPath:      windowsJoin(programData(), "DockSight", "logs", "agent.log"),
		Windows:      true,
	}
}

// ConfigPath is the agent configuration file.
//
// The separator comes from the layout, never from the host the CLI happens to
// be running on. These strings are written into a YAML file the agent reads
// back, so they have to be paths on the machine the agent will run on: a
// Linux target must get "/etc/docksight-agent/identity.json" even when a
// Windows-built CLI generated it, and a Windows target must get
// "C:\ProgramData\DockSight\agent\identity.json" rather than the mixed
// separators filepath would produce anywhere but Windows.
func (l Layout) ConfigPath() string {
	return l.join("config.yaml")
}

// IdentityPath is where the agent persists the identity the platform
// assigns it.
func (l Layout) IdentityPath() string {
	return l.join("identity.json")
}

// LogsCommand is how an operator follows the agent's output on this host.
func (l Layout) LogsCommand() string {

	if l.Windows {
		return "Get-Content '" + l.LogPath + "' -Tail 50 -Wait"
	}

	return "journalctl -u " + l.ServiceName + " -f"
}

// RestartCommand is how an operator restarts the agent on this host.
func (l Layout) RestartCommand() string {

	if l.Windows {
		return "Restart-Service " + l.ServiceName
	}

	return "sudo systemctl restart " + l.ServiceName
}

func (l Layout) join(name string) string {

	if l.Windows {
		return windowsJoin(strings.TrimRight(l.ConfigDir, `\/`), name)
	}

	return path.Join(l.ConfigDir, name)
}

// windowsJoin joins path elements the way Windows does, on any host.
// filepath.Join would use "/" everywhere but Windows, which is exactly the
// mixed-separator config this package must not generate.
func windowsJoin(elements ...string) string {
	return strings.Join(elements, `\`)
}

func programFiles() string {
	return environmentOr("ProgramFiles", `C:\Program Files`)
}

func programData() string {
	return environmentOr("ProgramData", `C:\ProgramData`)
}

func environmentOr(name string, fallback string) string {

	if value := strings.TrimRight(os.Getenv(name), `\`); value != "" {
		return value
	}

	return fallback
}

// agentEndpoint is the WebSocket path the platform exposes for agents.
const agentEndpoint = "agents"

// NormalizeServerURL turns whatever the operator typed into the WebSocket URL
// the agent connects to.
//
//	https://platform.example.com        -> wss://platform.example.com/agents
//	http://10.0.0.5:2002                -> ws://10.0.0.5:2002/agents
//	platform.example.com                -> wss://platform.example.com/agents
//	https://platform.example.com/       -> wss://platform.example.com/agents
//	wss://platform.example.com/agents   -> unchanged
//
// Slashes are joined with path.Join, so a trailing slash can never produce
// the "//agents" path that a naive concatenation would.
func NormalizeServerURL(raw string) (string, error) {

	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		return "", fmt.Errorf("the DockSight platform URL is required")
	}

	// A bare host has no scheme, and url.Parse would read "host:2002" as
	// scheme "host". Default to the secure scheme.
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)

	if err != nil {
		return "", fmt.Errorf("%q is not a valid URL: %w", raw, err)
	}

	scheme, err := websocketScheme(parsed.Scheme)

	if err != nil {
		return "", err
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("%q has no host", raw)
	}

	cleaned := path.Clean("/" + strings.Trim(parsed.Path, "/"))

	if cleaned == "/" {
		cleaned = ""
	}

	// Only append the endpoint when the operator has not already typed it.
	if !strings.HasSuffix(cleaned, "/"+agentEndpoint) {
		cleaned = path.Join(cleaned, agentEndpoint)
	}

	normalized := url.URL{
		Scheme: scheme,
		Host:   parsed.Host,
		Path:   "/" + strings.TrimPrefix(cleaned, "/"),
	}

	return normalized.String(), nil
}

// websocketScheme maps an entered scheme onto its WebSocket equivalent.
func websocketScheme(scheme string) (string, error) {

	switch strings.ToLower(scheme) {

	case "https", "wss":
		return "wss", nil

	case "http", "ws":
		return "ws", nil

	default:
		return "", fmt.Errorf(
			"unsupported scheme %q, use http, https, ws or wss",
			scheme,
		)
	}
}

// configTemplate matches the schema the agent parses in
// apps/agent/internal/config. Values are quoted so a URL or path containing
// YAML syntax cannot corrupt the file.
//
// The quoting is Go's, via %q, and that is safe for Windows paths as well:
// Go escapes a backslash as "\\", and a YAML double-quoted scalar unescapes
// it back to one. "C:\ProgramData\DockSight\agent" survives the round trip,
// and so does the named pipe "\\.\pipe\docker_engine". Single quotes would
// not need the escaping but would break on a path containing an apostrophe,
// which a Windows user profile can.
const configTemplate = `# DockSight Agent configuration
# Written by "docksight agent install". Edit and restart the service:
#   %s

agent:
  data_dir: %q
  identity_file: %q

server:
  # DockSight platform WebSocket endpoint for agent registration
  url: %q

docker:
  socket: %q

logging:
  level: %q
`

// RenderConfig builds the agent configuration file.
func RenderConfig(layout Layout, serverURL string) string {

	return fmt.Sprintf(
		configTemplate,
		layout.RestartCommand(),
		layout.ConfigDir,
		layout.IdentityPath(),
		serverURL,
		layout.DockerSocket,
		layout.LogLevel,
	)
}

// WriteConfig creates the configuration directory and writes config.yaml.
//
// The file is rewritten on every install so a changed platform URL takes
// effect, but identity.json is never touched: it carries the identity the
// platform issued, and replacing it would register the host a second time and
// orphan its history.
func WriteConfig(layout Layout, serverURL string) error {

	if err := os.MkdirAll(layout.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", layout.ConfigDir, err)
	}

	content := RenderConfig(layout, serverURL)

	// 0640: the file names the platform the host reports to, and later gains
	// enrollment credentials. Readable by root, not by the world.
	//
	// Windows ignores the mode and the file inherits the ACL of its parent.
	// That parent is under ProgramData, which grants ordinary users read but
	// not write — weaker than 0640, and the same protection every other
	// machine-wide service configuration on the host gets.
	if err := os.WriteFile(layout.ConfigPath(), []byte(content), 0o640); err != nil {
		return fmt.Errorf("failed to write %s: %w", layout.ConfigPath(), err)
	}

	return nil
}

// ReadServerURL returns the platform URL recorded in an existing
// installation, so an update does not require the operator to retype it.
//
// This reads the one key it needs rather than unmarshalling YAML: the CLI has
// no YAML dependency, and a config the agent accepts but this cannot parse
// would block updates for no good reason.
func ReadServerURL(layout Layout) (string, error) {

	file, err := os.Open(layout.ConfigPath())

	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	inServer := false

	for scanner.Scan() {

		line := scanner.Text()

		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// A key at column zero starts a new top-level section.
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inServer = strings.HasPrefix(trimmed, "server:")
			continue
		}

		if !inServer {
			continue
		}

		value, found := strings.CutPrefix(trimmed, "url:")

		if !found {
			continue
		}

		return strings.Trim(strings.TrimSpace(value), `"'`), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no server.url found in %s", layout.ConfigPath())
}

// UpdateServerURL changes only server.url in an existing config file. Other
// operator-managed settings are preserved, and identity.json is never opened.
func UpdateServerURL(layout Layout, serverURL string) error {

	content, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		return err
	}

	info, err := os.Stat(layout.ConfigPath())

	if err != nil {
		return err
	}

	updated, err := replaceServerURL(string(content), serverURL)

	if err != nil {
		return err
	}

	staged, err := os.CreateTemp(layout.ConfigDir, ".config-")

	if err != nil {
		return err
	}

	stagedPath := staged.Name()

	defer os.Remove(stagedPath)

	if _, err := staged.WriteString(updated); err != nil {
		staged.Close()
		return err
	}

	if err := staged.Close(); err != nil {
		return err
	}

	if err := os.Chmod(stagedPath, info.Mode().Perm()); err != nil {
		return err
	}

	if err := os.Rename(stagedPath, layout.ConfigPath()); err != nil {
		return fmt.Errorf("failed to update %s: %w", layout.ConfigPath(), err)
	}

	return nil
}

func replaceServerURL(content string, serverURL string) (string, error) {

	lines := strings.SplitAfter(content, "\n")
	inServer := false

	for index, rawLine := range lines {

		line := rawLine
		ending := ""

		switch {
		case strings.HasSuffix(line, "\r\n"):
			line = strings.TrimSuffix(line, "\r\n")
			ending = "\r\n"
		case strings.HasSuffix(line, "\n"):
			line = strings.TrimSuffix(line, "\n")
			ending = "\n"
		}

		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inServer = strings.HasPrefix(trimmed, "server:")
			continue
		}

		if !inServer {
			continue
		}

		if _, found := strings.CutPrefix(trimmed, "url:"); !found {
			continue
		}

		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		lines[index] = indent + "url: " + strconv.Quote(serverURL) + ending

		return strings.Join(lines, ""), nil
	}

	return "", fmt.Errorf("no server.url found in config")
}

// IdentityExists reports whether this host has already registered.
func IdentityExists(layout Layout) bool {

	info, err := os.Stat(layout.IdentityPath())

	return err == nil && info.Size() > 0
}
