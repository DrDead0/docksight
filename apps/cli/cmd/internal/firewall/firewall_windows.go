//go:build windows

package firewall

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// AllowPort opens an inbound TCP port, and reports whether it had to.
//
// netsh is used rather than the New-NetFirewallRule cmdlet because it needs
// no PowerShell and no execution policy, and because it is present on every
// Windows edition the platform could be installed on.
//
// Existence is checked first rather than deleting and re-adding. A blind
// re-add would produce a second rule with the same name on every install, and
// an operator who has deliberately narrowed the rule — scoped it to a subnet,
// say — would have that undone by a routine upgrade.
func AllowPort(ctx context.Context, port int) (bool, error) {

	exists, err := ruleExists(ctx)

	if err != nil {
		return false, err
	}

	if exists {
		return false, nil
	}

	command := exec.CommandContext(
		ctx,
		"netsh", "advfirewall", "firewall", "add", "rule",
		"name="+RuleName,
		"description="+Description(port),
		"dir=in",
		"action=allow",
		"protocol=TCP",
		"localport="+strconv.Itoa(port),
		// Private and domain networks only. A platform reachable from a
		// coffee-shop network is not what installing a monitoring tool asks
		// for, and an operator who does want that can widen the rule.
		"profile=domain,private",
	)

	output, err := command.CombinedOutput()

	if err != nil {
		return false, fmt.Errorf(
			"failed to open TCP port %d in Windows Firewall: %s",
			port,
			firstLine(output),
		)
	}

	return true, nil
}

// ruleExists reports whether the rule is already present.
func ruleExists(ctx context.Context) (bool, error) {

	command := exec.CommandContext(
		ctx,
		"netsh", "advfirewall", "firewall", "show", "rule",
		"name="+RuleName,
	)

	output, err := command.CombinedOutput()

	// netsh exits non-zero when no rule matches, which is an answer rather
	// than a failure. Anything else — netsh missing, the firewall service
	// stopped — would have failed before producing this output.
	if err != nil {
		return false, nil
	}

	return strings.Contains(string(output), RuleName), nil
}

func firstLine(output []byte) string {

	message := strings.TrimSpace(string(output))

	if message == "" {
		return "no output from netsh"
	}

	if index := strings.IndexAny(message, "\r\n"); index >= 0 {
		return message[:index]
	}

	return message
}
