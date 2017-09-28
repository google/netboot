package pixiecore

import (
	"go.universe.tf/netboot/dhcp6"
	"fmt"
)

func (s *ServerV6) serveDHCP(conn *dhcp6.Conn) error {
	s.log("dhcpv6", "Waiting for packets...\n")
	for {
		pkt, src, err := conn.RecvDHCP()
		if err != nil {
			return fmt.Errorf("Error receiving DHCP packet: %s", err)
		}
		if err := pkt.ShouldDiscard(s.Duid); err != nil {
			s.log("dhcpv6", fmt.Sprintf("Discarding (%d) packet (%d): %s\n", pkt.Type, pkt.TransactionID, err))
			continue
		}

		s.log("dhcpv6", fmt.Sprintf("Received (%d) packet (%d): %s\n", pkt.Type, pkt.TransactionID, pkt.Options.HumanReadable()))

		response, _ := pkt.BuildResponse(s.Duid)
		if err := conn.SendDHCP(src, response); err != nil {
			s.log("dhcpv6", fmt.Sprintf("Error sending reply (%d) (%d): %s", pkt.Type, pkt.TransactionID, err))
			continue
		}

		reply_packet, _ := dhcp6.MakePacket(response, len(response))
		reply_opts := reply_packet.Options

		s.log("dhcpv6", fmt.Sprintf("Sent (%d) packet (%d): %s\n", reply_packet.Type, reply_packet.TransactionID, reply_opts.HumanReadable()))
	}
}
