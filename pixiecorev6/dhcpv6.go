package pixiecorev6

import (
	"go.universe.tf/netboot/dhcp6"
	"fmt"
)

func (s *ServerV6) serveDHCP(conn *dhcp6.Conn, packetBuilder *dhcp6.PacketBuilder) error {
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

		response := packetBuilder.BuildResponse(pkt)

		marshalled_response, err := response.Marshal()
		if err != nil {
			s.log("dhcpv6", fmt.Sprintf("Error marshalling response: %s", response.Type, response.TransactionID, err))
			continue
		}

		if err := conn.SendDHCP(src, marshalled_response); err != nil {
			s.log("dhcpv6", fmt.Sprintf("Error sending reply (%d) (%d): %s", response.Type, response.TransactionID, err))
			continue
		}

		s.log("dhcpv6", fmt.Sprintf("Sent (%d) packet (%d): %s\n", response.Type, response.TransactionID, response.Options.HumanReadable()))
	}
}
