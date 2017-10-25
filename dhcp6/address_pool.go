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
	ReserveAddresses(clientId []byte, interfaceIds [][]byte) []*IdentityAssociation
	ReleaseAddresses(clientId []byte, interfaceIds [][]byte)
}
