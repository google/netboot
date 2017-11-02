package dhcp6

import (
	"net"
	"time"
)

// Associates an ip address with an individual network interface of a client
type IdentityAssociation struct {
	IPAddress   net.IP
	ClientID    []byte
	InterfaceID []byte
	CreatedAt   time.Time
}

// Keeps track of assigned and available ip address in an address pool
type AddressPool interface {
	ReserveAddresses(clientID []byte, interfaceIds [][]byte) ([]*IdentityAssociation, error)
	ReleaseAddresses(clientID []byte, interfaceIds [][]byte)
}
