package dhcp6

import "net"

// BootConfiguration implementation provides values for dhcp options served to dhcp clients
type BootConfiguration interface {
	GetBootURL(id []byte, clientArchType uint16) ([]byte, error)
	GetPreference() []byte
	GetRecursiveDNS() []net.IP
}
