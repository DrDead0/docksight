package system

import (
	"net"
	"strings"
)

// routeProbe is the address used to ask the routing table which local
// interface faces the network. It is a UDP "connection", so nothing is ever
// sent to it and the host does not have to be reachable — connecting a UDP
// socket only assigns a local address.
const routeProbe = "8.8.8.8:80"

// PrimaryIPv4 returns the address other machines should use to reach this
// host, falling back to "localhost" when there is no such address.
//
// The answer is taken from the routing table rather than by walking the
// interface list and returning the first candidate. That older approach was
// wrong on any host with virtual networking, and Windows hosts running the
// platform always have some: a machine here enumerates a VirtualBox
// host-only adapter on 192.168.56.1 and a WSL adapter on 172.31.208.1 before
// the Wi-Fi address that agents can actually reach. Printing either as the
// platform URL sends every agent at an address that does not answer, and the
// failure looks like a firewall problem rather than a wrong address.
//
// This does not prove reachability — a firewall or a NAT in front of the host
// can still block it — only that it is the address this host uses to talk to
// the wider network.
func PrimaryIPv4() string {

	if address := routedIPv4(); address != "" {
		return address
	}

	if address := firstGlobalIPv4(); address != "" {
		return address
	}

	return "localhost"
}

// routedIPv4 asks the kernel which local address would be used to reach the
// wider network.
func routedIPv4() string {

	connection, err := net.Dial("udp4", routeProbe)

	if err != nil {
		return ""
	}

	defer connection.Close()

	address, ok := connection.LocalAddr().(*net.UDPAddr)

	if !ok || address.IP == nil {
		return ""
	}

	ip := address.IP.To4()

	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return ""
	}

	return ip.String()
}

// firstGlobalIPv4 is the fallback for a host with no route to the wider
// network — an air-gapped install, or a laptop with every interface down.
// Virtual adapters are skipped by name where they can be recognised, which is
// a heuristic and is why this is the fallback and not the answer.
func firstGlobalIPv4() string {

	interfaces, err := net.Interfaces()

	if err != nil {
		return ""
	}

	for _, candidate := range interfaces {

		if candidate.Flags&net.FlagUp == 0 || candidate.Flags&net.FlagLoopback != 0 {
			continue
		}

		if isVirtualInterface(candidate.Name) {
			continue
		}

		addresses, err := candidate.Addrs()

		if err != nil {
			continue
		}

		for _, address := range addresses {

			var ip net.IP

			switch value := address.(type) {

			case *net.IPNet:
				ip = value.IP

			case *net.IPAddr:
				ip = value.IP
			}

			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			if ip = ip.To4(); ip == nil {
				continue
			}

			return ip.String()
		}
	}

	return ""
}

// virtualInterfaceMarkers name the adapters a container or VM host adds. They
// carry addresses that are real but reachable only from this machine.
var virtualInterfaceMarkers = []string{
	"vethernet", // Windows: Hyper-V, WSL, Docker Desktop
	"virtualbox",
	"vmware",
	"docker",
	"br-",   // docker bridge networks
	"veth",  // container side of a docker bridge
	"cni",   // kubernetes
	"tun",   // vpn
	"tap",   // vpn
	"utun",  // macos vpn
	"zt",    // zerotier
	"wg",    // wireguard
	"vmnet", // vmware host-only
}

func isVirtualInterface(name string) bool {

	lower := strings.ToLower(name)

	for _, marker := range virtualInterfaceMarkers {

		if strings.Contains(lower, marker) {
			return true
		}
	}

	return false
}
