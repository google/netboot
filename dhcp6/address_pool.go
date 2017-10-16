package dhcp6

import (
	"net"
	"time"
)

type IdentityAssociation struct {
	ipAddress	net.IP
	clientId	[]byte
	interfaceId	[]byte
	createdAt	time.Time
	t1			uint32
	t2			uint32
}

type AddressPool interface {
	ReserveAddress(clientId, interfaceId []byte) *IdentityAssociation
	ReleaseAddress(clientId, interfaceId []byte, addr net.IP)
}













