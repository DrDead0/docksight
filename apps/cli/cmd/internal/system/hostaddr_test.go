package system

import (
	"net"
	"testing"
)

func TestPrimaryIPv4ReturnsSomething(t *testing.T) {
	got := PrimaryIPv4()
	if got == "" {
		t.Fatal("PrimaryIPv4 returned empty string")
	}
	// Either a dotted IPv4 or the localhost fallback.
	if got != "localhost" && !looksLikeIPv4(got) {
		t.Fatalf("unexpected address %q", got)
	}
}

// The address is printed as the one agents should dial, so a loopback or an
// unroutable answer is worse than no answer: it looks like a working install
// and produces agents that never connect.
func TestPrimaryIPv4IsNotLoopback(t *testing.T) {

	got := PrimaryIPv4()

	if got == "localhost" {
		t.Skip("this host has no network address")
	}

	address := net.ParseIP(got)

	if address == nil {
		t.Fatalf("%q is not an IP address", got)
	}

	if address.IsLoopback() {
		t.Fatalf("%q is a loopback address", got)
	}

	if address.IsUnspecified() {
		t.Fatalf("%q is the unspecified address", got)
	}
}

// The address must belong to this machine, whichever route produced it.
func TestPrimaryIPv4BelongsToAnInterface(t *testing.T) {

	got := PrimaryIPv4()

	if got == "localhost" {
		t.Skip("this host has no network address")
	}

	addresses, err := net.InterfaceAddrs()

	if err != nil {
		t.Skip(err)
	}

	for _, address := range addresses {

		if network, ok := address.(*net.IPNet); ok && network.IP.String() == got {
			return
		}
	}

	t.Fatalf("%q is not an address of this host", got)
}

// Virtual adapters are the reason the routing table is consulted first. These
// names are all real: a Windows host running the platform typically carries
// several, and the old first-match-wins scan returned whichever the OS
// happened to enumerate first.
func TestIsVirtualInterface(t *testing.T) {

	virtual := []string{
		"vEthernet (WSL (Hyper-V firewall))",
		"vEthernet (Default Switch)",
		"VirtualBox Host-Only Network",
		"VMware Network Adapter VMnet1",
		"docker0",
		"br-1a2b3c4d",
		"veth0a1b2c3",
		"cni0",
		"tun0",
	}

	for _, name := range virtual {

		if !isVirtualInterface(name) {
			t.Errorf("%q was not recognised as virtual", name)
		}
	}

	physical := []string{
		"eth0",
		"ens18",
		"enp0s31f6",
		"wlan0",
		"Wi-Fi",
		"Ethernet 2",
	}

	for _, name := range physical {

		if isVirtualInterface(name) {
			t.Errorf("%q was wrongly treated as virtual", name)
		}
	}
}

func looksLikeIPv4(s string) bool {
	// Minimal shape check: four dotted decimal groups.
	dots := 0
	for _, c := range s {
		if c == '.' {
			dots++
		}
	}
	return dots == 3
}
