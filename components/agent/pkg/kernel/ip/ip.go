package ip

import (
	"fmt"
	"net"
	"strings"
)

type NetConfig struct {
	clientIp       string
	serverIp       string
	gatewayIp      string
	netmask        string
	hostname       string
	device         string
	autoConf       string
	primaryDnsIp   string
	secondaryDnsIp string
	ntpServerIp    string
}

func ParseFromString(input string) (*NetConfig, error) {
	parts := strings.Split(input, ":")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid input string: %v", input)
	}
	return &NetConfig{
		clientIp:       string(parts[0]),
		serverIp:       string(parts[1]),
		gatewayIp:      string(parts[2]),
		netmask:        string(parts[3]),
		hostname:       string(parts[4]),
		device:         string(parts[5]),
		autoConf:       string(parts[6]),
		primaryDnsIp:   string(parts[7]),
		secondaryDnsIp: string(parts[8]),
		ntpServerIp:    string(parts[9]),
	}, nil
}

func (nc *NetConfig) ClientIp() net.IP {
	return net.ParseIP(nc.clientIp)
}

func (nc *NetConfig) ServerIp() net.IP {
	return net.ParseIP(nc.serverIp)
}

func (nc *NetConfig) Gateway() net.IP {
	return net.ParseIP(nc.gatewayIp)
}

func (nc *NetConfig) Netmask() net.IPMask {
	return net.IPMask(net.ParseIP(nc.netmask).To4())
}

func (nc *NetConfig) Hostname() string {
	return nc.hostname
}

func (nc *NetConfig) Device() string {
	return nc.device
}

func (nc *NetConfig) AutoConf() string {
	return nc.autoConf
}

func (nc *NetConfig) PrimaryDns() net.IP {
	return net.ParseIP(nc.primaryDnsIp)
}

func (nc *NetConfig) SecondaryDns() net.IP {
	return net.ParseIP(nc.secondaryDnsIp)
}

func (nc *NetConfig) NtpServer() net.IP {
	return net.ParseIP(nc.ntpServerIp)
}
