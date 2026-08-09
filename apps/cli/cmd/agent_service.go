package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/agent/install"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/systemd"
	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/ui"

	"github.com/spf13/cobra"
)

// How long to wait for systemd to report the unit active after start/restart.
// systemctl start returning 0 only means the request was accepted.
const agentServiceSettle = 5 * time.Second
const agentServicePoll = time.Second

const agentServiceRootHint = "this command must be run as root — try `sudo docksight agent %s`"

func agentUnitName() string {
	return install.DefaultLayout().ServiceName
}

func requireRoot(command string) error {
	if os.Geteuid() == 0 {
		return nil
	}
	return fmt.Errorf(agentServiceRootHint, command)
}

func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "permission denied") ||
		strings.Contains(message, "access denied") ||
		strings.Contains(message, "interactive authentication required")
}

func wrapAgentServiceError(command string, err error) error {
	if err == nil {
		return nil
	}
	if isPermissionError(err) {
		return fmt.Errorf(agentServiceRootHint, command)
	}
	return err
}

// waitActive polls until the unit reports active or the settle window passes.
func waitActive(ctx context.Context, manager *systemd.Manager, unit string) (bool, error) {
	deadline := time.Now().Add(agentServiceSettle)

	for {
		active, err := manager.IsActive(ctx, unit)
		if err != nil {
			return false, err
		}
		if active {
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		timer := time.NewTimer(agentServicePoll)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false, ctx.Err()
		case <-timer.C:
		}
	}
}

var agentStartCMD = &cobra.Command{
	Use:   "start",
	Short: "Start the agent service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot("start"); err != nil {
			return err
		}

		unit := agentUnitName()
		manager := systemd.NewManager()
		ui.Info("Starting " + unit)

		if err := wrapAgentServiceError("start", manager.Start(cmd.Context(), unit)); err != nil {
			return err
		}

		active, err := waitActive(cmd.Context(), manager, unit)
		if err != nil {
			return wrapAgentServiceError("start", err)
		}
		if !active {
			return fmt.Errorf(
				"%s did not become active — inspect with: journalctl -u %s -n 50",
				unit,
				unit,
			)
		}

		ui.Success(unit + " is active")
		return nil
	},
}

var agentStopCMD = &cobra.Command{
	Use:   "stop",
	Short: "Stop the agent service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot("stop"); err != nil {
			return err
		}

		unit := agentUnitName()
		manager := systemd.NewManager()
		ui.Info("Stopping " + unit)

		if err := wrapAgentServiceError("stop", manager.Stop(cmd.Context(), unit)); err != nil {
			return err
		}

		ui.Success(unit + " stopped")
		return nil
	},
}

var agentRestartCMD = &cobra.Command{
	Use:   "restart",
	Short: "Restart the agent service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot("restart"); err != nil {
			return err
		}

		unit := agentUnitName()
		manager := systemd.NewManager()
		ui.Info("Restarting " + unit)

		if err := wrapAgentServiceError("restart", manager.Restart(cmd.Context(), unit)); err != nil {
			return err
		}

		active, err := waitActive(cmd.Context(), manager, unit)
		if err != nil {
			return wrapAgentServiceError("restart", err)
		}
		if !active {
			return fmt.Errorf(
				"%s did not become active after restart — inspect with: journalctl -u %s -n 50",
				unit,
				unit,
			)
		}

		ui.Success(unit + " is active")
		return nil
	},
}

func init() {
	agentCMD.AddCommand(agentStartCMD)
	agentCMD.AddCommand(agentStopCMD)
	agentCMD.AddCommand(agentRestartCMD)
}
