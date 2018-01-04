package pixiecore

import (
	"go.universe.tf/netboot/dhcp6"
	"fmt"
	"time"
	"encoding/binary"
	"net"
)

type ServerV6 struct {
	Address    		string
	Port       		string
	Duid       		[]byte

	BootConfig		dhcp6.BootConfiguration
	PacketBuilder	*dhcp6.PacketBuilder
	AddressPool 	dhcp6.AddressPool

	errs chan error

	Log func(subsystem, msg string)
	Debug func(subsystem, msg string)
}

func NewServerV6() *ServerV6 {
	ret := &ServerV6{
		Port: "547",
	}
	return ret
}

func (s *ServerV6) Serve() error {
	s.log("dhcp", "starting...")

	dhcp, err := dhcp6.NewConn(s.Address, s.Port)
	if err != nil {
		return err
	}

	s.debug("dhcp", "new connection...")

	// 5 buffer slots, one for each goroutine, plus one for
	// Shutdown(). We only ever pull the first error out, but shutdown
	// will likely generate some spurious errors from the other
	// goroutines, and we want them to be able to dump them without
	// blocking.
	s.errs = make(chan error, 6)

	s.SetDUID(dhcp.SourceHardwareAddress())

	go func() { s.errs <- s.serveDHCP(dhcp) }()

	// Wait for either a fatal error, or Shutdown().
	err = <-s.errs
	dhcp.Close()

	s.log("dhcp", "stopped...")
	return err
}

// Shutdown causes Serve() to exit, cleaning up behind itself.
func (s *ServerV6) Shutdown() {
	select {
	case s.errs <- nil:
	default:
	}
}

func (s *ServerV6) log(subsystem, format string, args ...interface{}) {
	if s.Log == nil {
		return
	}
	s.Log(subsystem, fmt.Sprintf(format, args...))
}

func (s *ServerV6) debug(subsystem, format string, args ...interface{}) {
	if s.Debug == nil {
		return
	}
	s.Debug(subsystem, fmt.Sprintf(format, args...))
}

func (s *ServerV6) SetDUID(addr net.HardwareAddr) {
	duid := make([]byte, len(addr) + 8) // see rfc3315, section 9.2, DUID-LT

	copy(duid[0:], []byte{0, 1}) //fixed, x0001
	copy(duid[2:], []byte{0, 1}) //hw type ethernet, x0001

	utcLoc, _ := time.LoadLocation("UTC")
	sinceJanFirst2000 := time.Since(time.Date(2000, time.January, 1, 0, 0, 0, 0, utcLoc))
	binary.BigEndian.PutUint32(duid[4:], uint32(sinceJanFirst2000.Seconds()))

	copy(duid[8:], addr)

	s.Duid = duid
}