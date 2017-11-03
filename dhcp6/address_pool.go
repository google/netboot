package dhcp6

import (
	"net"
	"time"
)

// IdentityAssociation associates an ip address with a network interface of a client
type IdentityAssociation struct {
	IPAddress   net.IP
	ClientID    []byte
	InterfaceID []byte
	CreatedAt   time.Time
}

// AddressPool keeps track of assigned and available ip address in an address pool
type AddressPool interface {
	ReserveAddresses(clientID []byte, interfaceIds [][]byte) ([]*IdentityAssociation, error)
	ReleaseAddresses(clientID []byte, interfaceIds [][]byte)
}
