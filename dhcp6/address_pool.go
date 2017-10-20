package dhcp6

import (
	"net"
	"time"
)

type IdentityAssociation struct {
	IpAddress   net.IP
	ClientId    []byte
	InterfaceId []byte
	CreatedAt   time.Time
}

type AddressPool interface {
	ReserveAddress(clientId, interfaceId []byte) *IdentityAssociation
	ReleaseAddress(clientId, interfaceId []byte)
}
