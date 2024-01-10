package network

import (
	"fmt"
	"log"
	"net"

	"apollo/agent/pkg/kernel/ip"

	"github.com/vishvananda/netlink"
)

func Apply(config *ip.NetConfig) error {
	// Find the network interface by name.
	link, err := netlink.LinkByName(config.Device())
	if err != nil {
		return fmt.Errorf("failed to find network interface: %v", err)
	}

	// Configure the IP address for the interface.
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: config.ClientIp(), Mask: config.Netmask()}}
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
		Gw:        config.Gateway(),
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("failed to configure default gateway: %v", err)
	}

	log.Printf("Network interface %s configured successfully", config.Device())
	return nil
}
