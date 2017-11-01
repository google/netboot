package dhcp6

import (
	"net"
	"time"
)

type IdentityAssociation struct {
	IPAddress   net.IP
	ClientID    []byte
	InterfaceID []byte
	CreatedAt   time.Time
}

type AddressPool interface {
	ReserveAddresses(clientID []byte, interfaceIds [][]byte) ([]*IdentityAssociation, error)
	ReleaseAddresses(clientID []byte, interfaceIds [][]byte)
}
