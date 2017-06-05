package dhcp6

import (
	"io"
	"net"
	"time"
	"golang.org/x/net/ipv6"
	"fmt"
)

type conn interface {
	io.Closer
	Recv([]byte) (b []byte, addr *net.UDPAddr, ifidx int, err error)
	Send(b []byte, addr *net.UDPAddr, ifidx int) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type Conn struct {
	conn *ipv6.PacketConn
	group net.IP
	ifi *net.Interface
	listenAddress string
	listenPort string
}

func NewConn(addr string) (*Conn, error) {
	ifi, err := InterfaceIndexByAddress(addr)
	if err != nil {
		return nil, err
	}

	group := net.ParseIP("ff02::1:2")
	c, err := net.ListenPacket("udp6", "[::]:547")
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
		listenPort: "547",
	}, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func InterfaceIndexByAddress(ifAddr string) (*net.Interface, error) {
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
			if addr.String() == ifAddr {
				return &ifi, nil
			}
		}
	}
	return nil, fmt.Errorf("Couldn't find an interface with address %s", ifAddr)
}

func (c *Conn) RecvDHCP() (*Packet, net.IP, error) {
	b := make([]byte, 1500)
	for {
		packetSize, rcm, _, err := c.conn.ReadFrom(b)
		if err != nil {
			return nil, nil, err
		}
		if c.ifi.Index != 0 && rcm.IfIndex != c.ifi.Index {
			continue
		}
		if !rcm.Dst.IsMulticast() || !rcm.Dst.Equal(c.group) {
			continue // unknown group, discard
		}
		pkt := MakePacket(b, packetSize)

		return pkt, rcm.Src, nil
	}
}

func (c *Conn) SendDHCP(dst net.IP, p []byte) error {
	dstAddr, err := net.ResolveUDPAddr("udp6", fmt.Sprintf("[%s]:%s", dst.String() + "%en0", "546"))
	if err != nil {
		return fmt.Errorf("Error resolving ipv6 address %s: %s", dst.String(), err)
	}
	_, err = c.conn.WriteTo(p, nil, dstAddr)
	if err != nil {
		return fmt.Errorf("Error sending a reply to %s: %s", dst.String(), err)
	}
	return nil
}

func (c *Conn) SourceHardwareAddress() net.HardwareAddr {
	return c.ifi.HardwareAddr
}