package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/vishvananda/netlink"
)

type ipConfiguration struct {
	// IP address of the client.
	IP net.IP

	// Netmask for local network interface.
	Netmask net.IPMask

	// Gateway defines the gateway ip address.
	Gateway net.IP

	// Name of network device to use.
	Device string

	// IP adresses for the primary and secondary nameserver.
	Nameservers []net.IP
}

func ReadKernelArgIP() (*ipConfiguration, error) {
	cmdline, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/cmdline: %v", err)
	}

	kernelArgs := string(cmdline)

	for _, arg := range strings.Fields(kernelArgs) {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "ip":
			return ParseIPConfiguration(value)
		}
	}

	return nil, fmt.Errorf("failed to read kernel arg 'ip' from cmdline: %q", kernelArgs)
}

func (i *ipConfiguration) Apply() error {
	// Find the network interface by name.
	link, err := netlink.LinkByName(i.Device)
	if err != nil {
		return fmt.Errorf("failed to find network interface: %v", err)
	}

	// Configure the IP address for the interface.
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: i.IP, Mask: i.Netmask}}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("failed to configure IP address: %v", err)
	}

	// Bring the interface up.
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring up the interface: %v", err)
	}

	// Configure the default gateway.
	defaultRoute := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        i.Gateway,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("failed to configure default gateway: %v", err)
	}

	log.Printf("Network interface %s configured successfully", i.Device)
	return nil
}

func ParseIPConfiguration(ipConfig string) (*ipConfiguration, error) {
	parts := strings.Split(ipConfig, ":")
	if len(parts) < 1 {
		return nil, fmt.Errorf("failed to split ip configuration string: %q", ipConfig)
	}

	// Reference to kernel args:
	// https://www.kernel.org/doc/Documentation/filesystems/nfs/nfsroot.txt

	// IP address of the client.
	clientIP := string(parts[0])

	// IP address of a gateway if the server is on a different subnet.
	gatewayIP := string(parts[2])

	// Netmask for local network interface.
	netmask := string(parts[3])

	// Name of network device to use.
	device := string(parts[5])

	// IP address of primary nameserver.
	dns0IP := string(parts[7])

	// IP address of secondary nameserver.
	dns1IP := string(parts[8])

	ipConfiguration := ipConfiguration{
		IP:          net.ParseIP(clientIP),
		Netmask:     net.IPMask(net.ParseIP(netmask).To4()),
		Gateway:     net.ParseIP(gatewayIP),
		Device:      device,
		Nameservers: []net.IP{net.ParseIP(dns0IP), net.ParseIP(dns1IP)},
	}

	return &ipConfiguration, nil
}
