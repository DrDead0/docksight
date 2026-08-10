// Package firewall opens the port agents connect to.
//
// It exists for Windows, where the default inbound policy is to block. The
// platform publishes its dashboard and the agent WebSocket endpoint on one
// host port, and without a rule the port is reachable from the machine itself
// and from nowhere else — which looks exactly like a broken agent rather than
// a closed firewall, and is the failure this package exists to prevent.
//
// Linux hosts are left alone. Distributions disagree on whether anything is
// filtering at all (ufw, firewalld, plain nftables, nothing), an installer
// that guesses wrong either fails or silently writes a rule into a firewall
// that is not the active one, and a server administrator expects to make that
// decision themselves.
package firewall

import "fmt"

// RuleName is how the rule is identified, and how a second install recognises
// the rule a first one created rather than adding a duplicate.
const RuleName = "DockSight Platform"

// Description explains the rule to whoever finds it later in the firewall
// console and wonders whether it is safe to remove.
func Description(port int) string {

	return fmt.Sprintf(
		"Allows DockSight agents and browsers to reach the DockSight platform on TCP port %d.",
		port,
	)
}
