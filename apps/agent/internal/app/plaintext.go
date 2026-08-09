package app

import (
	"net"
	"net/url"
	"strings"

	"docksight-agent/internal/logger"
)

// warnIfPlaintextServerURL logs when the agent will connect over ws:// to a
// non-loopback host. Local development (localhost / 127.0.0.1 / ::1) is quiet
// on purpose — there is no network hop to eavesdrop on.
func warnIfPlaintextServerURL(raw string) {
	u, err := url.Parse(raw)
	if err != nil {
		return
	}
	if !strings.EqualFold(u.Scheme, "ws") {
		return
	}
	if isLoopbackHost(u.Hostname()) {
		return
	}
	logger.Warn(
		"connecting over an unencrypted connection; container names, logs and command results will be readable on the network. Use a wss:// URL.",
	)
}

// isLoopbackHost reports whether host is a loopback name or address.
// Hostname lookalikes (e.g. localhost.evil.com) are not treated as loopback.
func isLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
