package system

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// supportedArchitectures are the platforms DockSight publishes builds for.
var supportedArchitectures = []string{"amd64", "arm64"}

// connectivityProbe is the host the installer must reach to download
// releases. Checking anything else would prove the wrong thing.
const connectivityProbe = "https://api.github.com"

// agentOperatingSystems are the hosts the agent installs on. The platform is
// not in this list: it is a Compose stack that still assumes Linux paths.
var agentOperatingSystems = []string{"linux", "windows"}

// NotElevatedError reports that this process lacks the privileges to install
// a service.
//
// It is a distinct type so a caller can recognise the one validation failure
// that ElevationHint answers, without matching on the message. Every other
// failed check needs a different fix — a Docker daemon that is down is not
// started by running as root.
type NotElevatedError struct {
	// Reason says what is missing, in the host's own terms.
	Reason string
}

func (e *NotElevatedError) Error() string {
	return e.Reason
}

// CheckOS reports whether this is a supported operating system for the
// platform install.
func CheckOS() error {

	if runtime.GOOS != "linux" {
		return fmt.Errorf(
			"DockSight installs on Linux only, this host runs %s",
			runtime.GOOS,
		)
	}

	return nil
}

// CheckAgentOS reports whether the agent installs on this operating system.
//
// This gate is wider than CheckOS on purpose. The agent is a single binary
// supervised by whatever service manager the host already runs — systemd on
// Linux, the Service Control Manager on Windows — so the only thing it needs
// from the OS is one it has an implementation for.
func CheckAgentOS() error {

	for _, supported := range agentOperatingSystems {

		if runtime.GOOS == supported {
			return nil
		}
	}

	return fmt.Errorf(
		"the DockSight Agent installs on %s only, this host runs %s",
		strings.Join(agentOperatingSystems, " and "),
		runtime.GOOS,
	)
}

// CheckArchitecture reports whether builds exist for this CPU.
func CheckArchitecture() error {

	for _, supported := range supportedArchitectures {

		if runtime.GOARCH == supported {
			return nil
		}
	}

	return fmt.Errorf(
		"unsupported architecture %s, DockSight publishes builds for %s",
		runtime.GOARCH,
		strings.Join(supportedArchitectures, " and "),
	)
}

// CheckDockerInstalled reports whether the docker CLI is on PATH.
func CheckDockerInstalled(ctx context.Context) error {

	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is not installed: %w", err)
	}

	return nil
}

// CheckDockerRunning reports whether the daemon answers. `docker info` fails
// both when the daemon is down and when the user cannot reach its socket, so
// the error distinguishes the two.
func CheckDockerRunning(ctx context.Context) error {

	command := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ServerVersion}}")

	output, err := command.CombinedOutput()

	if err == nil {
		return nil
	}

	message := strings.TrimSpace(string(output))

	if strings.Contains(message, "permission denied") {
		return fmt.Errorf(
			"cannot reach the Docker daemon: permission denied. Run as root, or add this user to the docker group",
		)
	}

	return fmt.Errorf("the Docker daemon is not running: %s", firstLine(message))
}

// CheckDockerCompose reports whether the compose plugin is available.
func CheckDockerCompose(ctx context.Context) error {

	command := exec.CommandContext(ctx, "docker", "compose", "version")

	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf(
			"docker compose is not available: %s",
			firstLine(strings.TrimSpace(string(output))),
		)
	}

	return nil
}

// CheckSystemd reports whether systemd is the init system. The presence of
// systemctl is not enough: it is installed inside containers and on hosts
// running a different init, where enabling a unit silently does nothing.
//
// Nothing calls this on Windows — see agent_windows.go for what takes its
// place there.
func CheckSystemd() error {

	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemd is not available: systemctl not found")
	}

	// systemd creates this directory only when it is PID 1.
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return fmt.Errorf(
			"systemd is not the init system on this host, the agent service cannot be installed",
		)
	}

	return nil
}

// CheckInternet reports whether the release host is reachable.
func CheckInternet(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodHead, connectivityProbe, nil)

	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {

		var dnsError *net.DNSError

		if errors.As(err, &dnsError) {
			return fmt.Errorf("no internet connectivity: cannot resolve %s", dnsError.Name)
		}

		return fmt.Errorf("no internet connectivity: cannot reach %s: %w", connectivityProbe, err)
	}

	defer response.Body.Close()

	return nil
}

func firstLine(message string) string {

	if message == "" {
		return "no output from docker"
	}

	if index := strings.IndexByte(message, '\n'); index >= 0 {
		return message[:index]
	}

	return message
}
