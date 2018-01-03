package dhcp6

import (
	"net"
	"golang.org/x/net/ipv6"
	"fmt"
)

// Conn is dhcpv6-specific socket
type Conn struct {
	conn *ipv6.PacketConn
	group net.IP
	ifi *net.Interface
	listenAddress string
	listenPort string
}

// NewConn creates a new Conn bound to specified address and port
func NewConn(addr, port string) (*Conn, error) {
	ifi, err := InterfaceByAddress(addr)
	if err != nil {
		return nil, err
	}

	group := net.ParseIP("ff02::1:2")
	c, err := net.ListenPacket("udp6", "[::]:" + port)
	if err != nil {
		return nil, err
	}
	pc := ipv6.NewPacketConn(c)
	if err := pc.JoinGroup(ifi, &net.UDPAddr{IP: group}); err != nil {
		pc.Close()
		return nil, err
	}

	if err := pc.SetControlMessage(ipv6.FlagSrc | ipv6.FlagDst, true); err != nil {
		pc.Close()
		return nil, err
	}

	return &Conn{
		conn:    pc,
		group: 	 group,
		ifi: ifi,
		listenAddress: addr,
		listenPort: port,
	}, nil
}

// Close closes Conn
func (c *Conn) Close() error {
	return c.conn.Close()
}

// InterfaceByAddress finds the interface bound to an ip address, or returns an error if none were found
func InterfaceByAddress(ifAddr string) (*net.Interface, error) {
	allIfis, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Error getting network interface information: %s", err)
	}
	for _, ifi := range allIfis {
		addrs, err := ifi.Addrs()
		if err != nil {
			return nil, fmt.Errorf("Error getting network interface address information: %s", err)
		}
		for _, addr := range addrs {
			if addrToIP(addr).String() == ifAddr {
				return &ifi, nil
			}
		}
	}
	return nil, fmt.Errorf("Couldn't find an interface with address %s", ifAddr)
}

func addrToIP(a net.Addr) net.IP {
	var ip net.IP
	switch v := a.(type) {
	case *net.IPAddr:
		ip = v.IP
	case *net.IPNet:
		ip = v.IP
	}

	return ip
}

// RecvDHCP reads next available dhcp packet from Conn
func (c *Conn) RecvDHCP() (*Packet, net.IP, error) {
	b := make([]byte, 1500)
	for {
		n, rcm, _, err := c.conn.ReadFrom(b)
		if err != nil {
			return nil, nil, err
		}
		if c.ifi.Index != 0 && rcm.IfIndex != c.ifi.Index {
			continue
		}
		if !rcm.Dst.IsMulticast() || !rcm.Dst.Equal(c.group) {
			continue // unknown group, discard
		}
		pkt, err := MakePacket(b, n)
		if err != nil {
			return nil, nil, err
		}

		return pkt, rcm.Src, nil
	}
}

// SendDHCP sends a dhcp packet to the specified ip address using Conn
func (c *Conn) SendDHCP(dst net.IP, p []byte) error {
	dstAddr := &net.UDPAddr{
		IP: dst,
		Port: 546,
	}
	_, err := c.conn.WriteTo(p, nil, dstAddr); if err != nil {
		return fmt.Errorf("Error sending a reply to %s: %s", dst.String(), err)
	}
	return nil
}

// SourceHardwareAddress returns hardware address of the interface used by Conn
func (c *Conn) SourceHardwareAddress() net.HardwareAddr {
	return c.ifi.HardwareAddr
}