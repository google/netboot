package dhcp6

import (
	"fmt"
	"net"
	"encoding/binary"
	"bytes"
	"golang.org/x/tools/go/gcimporter15/testdata"
)

type MessageType uint8

const (
	MsgSolicit MessageType = iota + 1
	MsgAdvertise
	MsgRequest
	MsgConfirm
	MsgRenew
	MsgRebind
	MsgReply
	MsgRelease
	MsgDecline
	MsgReconfigure
	MsgInformationRequest
	MsgRelayForw
	MsgRelayRepl
)

type Packet struct {
	Type          MessageType
	TransactionID [3]byte
	Options       Options
}

func MakePacket(bs []byte, len int) (*Packet, error) {
	options, err := MakeOptions(bs[4:]) // 4:len?
	if err != nil {
		return nil, fmt.Errorf("packet has malformed options section: %s", err)
	}
	ret := &Packet{Type: MessageType(bs[0]), Options: options}
	copy(ret.TransactionID[:], bs[1:4])
	return ret, nil
}

func (p *Packet) BuildResponse(serverDuid []byte) ([]byte, error) {
	switch p.Type {
	case MsgSolicit:
		return p.BuildMsgAdvertise(serverDuid)
	case MsgRequest:
		return p.BuildMsgReply(serverDuid)
	case MsgInformationRequest:
		return p.BuildMsgInformationRequestReply(serverDuid)
	case MsgRelease:
		return p.BuildMsgReleaseReply(serverDuid)
	default:
		return nil, nil
	}
}

func (p *Packet) BuildMsgAdvertise(serverDuid []byte) ([]byte, error) {
	in_options := p.Options
	ret_options := make(Options)

	ret_options.AddOption(&Option{Id: OptClientId, Length: uint16(len(in_options[OptClientId].Value)), Value: in_options[OptClientId].Value})
	ret_options.AddOption(MakeIaNaOption(in_options[OptIaNa].Value[0:4], 0, 0,
			MakeIaAddrOption(net.ParseIP("2001:db8:f00f:cafe::99"), 27000, 43200)))
	ret_options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverDuid)), Value: serverDuid})

	if 0x10 == binary.BigEndian.Uint16(in_options[OptClientArchType].Value) { // HTTPClient
		ret_options.AddOption(&Option{Id: OptVendorClass, Length: 16, Value: []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116}}) // HTTPClient
		ret_options.AddOption(&Option{Id: OptBootfileUrl, Length: 42, Value: []byte("http://[2001:db8:f00f:cafe::4]/bootx64.efi")})
	} else {
		ret_options.AddOption(&Option{Id: OptBootfileUrl, Length: 42, Value: []byte("http://[2001:db8:f00f:cafe::4]/script.ipxe")})
	}
//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	//ret_options.AddOption(OptBootfileParam, []byte("http://")
	//ret.Options[OptPreference] = [][]byte("http://")

	marshalled_ret_options, _ := ret_options.Marshal()

	ret := make([]byte, len(marshalled_ret_options) + 4, len(marshalled_ret_options) + 4)
	ret[0] = byte(MsgAdvertise)
	copy(ret[1:], p.TransactionID[:])
	copy(ret[4:], marshalled_ret_options)
	return ret, nil
}

// TODO: OptClientArchType may not be present

func (p *Packet) BuildMsgReply(serverDuid []byte) ([]byte, error) {
	in_options := p.Options
	ret_options := make(Options)

	ret_options.AddOption(&Option{Id: OptClientId, Length: uint16(len(in_options[OptClientId].Value)), Value: in_options[OptClientId].Value})
	ret_options.AddOption(MakeIaNaOption(in_options[OptIaNa].Value[0:4], 0, 0,
		MakeIaAddrOption(net.ParseIP("2001:db8:f00f:cafe::99"), 27000, 43200)))
	ret_options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverDuid)), Value: serverDuid})
	//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	if 0x10 == binary.BigEndian.Uint16(in_options[OptClientArchType].Value) { // HTTPClient
		ret_options.AddOption(&Option{Id: OptVendorClass, Length: 16, Value: []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116}}) // HTTPClient
		ret_options.AddOption(&Option{Id: OptBootfileUrl, Length: 42, Value: []byte("http://[2001:db8:f00f:cafe::4]/bootx64.efi")})
	} else {
		ret_options.AddOption(&Option{Id: OptBootfileUrl, Length: 42, Value: []byte("http://[2001:db8:f00f:cafe::4]/script.ipxe")})
	}
	marshalled_ret_options, _ := ret_options.Marshal()

	ret := make([]byte, len(marshalled_ret_options) + 4, len(marshalled_ret_options) + 4)
	ret[0] = byte(MsgReply)
	copy(ret[1:], p.TransactionID[:])
	copy(ret[4:], marshalled_ret_options)
	return ret, nil
}

func (p *Packet) BuildMsgInformationRequestReply(serverDuid []byte) ([]byte, error) {
	in_options := p.Options
	ret_options := make(Options)

	ret_options.AddOption(&Option{Id: OptClientId, Length: uint16(len(in_options[OptClientId].Value)), Value: in_options[OptClientId].Value})
	ret_options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverDuid)), Value: serverDuid})
	//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	if 0x10 == binary.BigEndian.Uint16(in_options[OptClientArchType].Value) { // HTTPClient
		ret_options.AddOption(&Option{Id: OptVendorClass, Length: 16, Value: []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116}}) // HTTPClient
		ret_options.AddOption(&Option{Id: OptBootfileUrl, Length: 42, Value: []byte("http://[2001:db8:f00f:cafe::4]/bootx64.efi")})
	} else {
		ret_options.AddOption(&Option{Id: OptBootfileUrl, Length: 42, Value: []byte("http://[2001:db8:f00f:cafe::4]/script.ipxe")})
	}
	marshalled_ret_options, _ := ret_options.Marshal()

	ret := make([]byte, len(marshalled_ret_options) + 4, len(marshalled_ret_options) + 4)
	ret[0] = byte(MsgReply)
	copy(ret[1:], p.TransactionID[:])
	copy(ret[4:], marshalled_ret_options)
	return ret, nil
}

func (p *Packet) BuildMsgReleaseReply(serverDuid []byte) ([]byte, error){
	in_options := p.Options
	ret_options := make(Options)

	ret_options.AddOption(&Option{Id: OptClientId, Length: uint16(len(in_options[OptClientId].Value)), Value: in_options[OptClientId].Value})
	ret_options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverDuid)), Value: serverDuid})
	v := make([]byte, 19, 19)
	copy(v[2:], []byte("Release received."))
	ret_options.AddOption(&Option{Id: OptStatusCode, Length: uint16(len(v)), Value: v})
	marshalled_ret_options, _ := ret_options.Marshal()

	ret := make([]byte, len(marshalled_ret_options) + 4, len(marshalled_ret_options) + 4)
	ret[0] = byte(MsgReply)
	copy(ret[1:], p.TransactionID[:])
	copy(ret[4:], marshalled_ret_options)
	//copy(ret.Options, marshalled_ret_options)
	return ret, nil
}

func (p *Packet) ShouldDiscard(serverDuid []byte) error {
	switch p.Type {
	case MsgSolicit:
		return ShouldDiscardSolicit(p)
	case MsgRequest:
		return ShouldDiscardRequest(p, serverDuid)
	case MsgInformationRequest:
		return ShouldDiscardInformationRequest(p, serverDuid)
	case MsgRelease:
		return nil // FIX ME!
	default:
		return fmt.Errorf("Unknown packet")
	}
}

func ShouldDiscardSolicit(p *Packet) error {
	options := p.Options
	if !options.RequestedBootFileUrlOption() {
		return fmt.Errorf("'Solicit' packet doesn't have file url option")
	}
	if !options.HasClientId() {
		return fmt.Errorf("'Solicit' packet has no client id option")
	}
	if options.HasServerId() {
		return fmt.Errorf("'Solicit' packet has server id option")
	}
	return nil
}

func ShouldDiscardRequest(p *Packet, serverDuid []byte) error {
	options := p.Options
	if !options.RequestedBootFileUrlOption() {
		return fmt.Errorf("'Request' packet doesn't have file url option")
	}
	if !options.HasClientId() {
		return fmt.Errorf("'Request' packet has no client id option")
	}
	if !options.HasServerId() {
		return fmt.Errorf("'Request' packet has no server id option")
	}
	if bytes.Compare(options[OptServerId].Value, serverDuid) != 0 {
		return fmt.Errorf("'Request' packet's server id option (%d) is different from ours (%d)", options[OptServerId].Value, serverDuid)
	}
	return nil
}

func ShouldDiscardInformationRequest(p *Packet, serverDuid []byte) error {
	options := p.Options
	if !options.RequestedBootFileUrlOption() {
		return fmt.Errorf("'Information-request' packet doesn't have boot file url option")
	}
	if options.HasIaNa() || options.HasIaTa() {
		return fmt.Errorf("'Information-request' packet has an IA option present")
	}
	if options.HasServerId() && (bytes.Compare(options[OptServerId].Value, serverDuid) != 0) {
		return fmt.Errorf("'Information-request' packet's server id option (%d) is different from ours (%d)", options[OptServerId].Value, serverDuid)
	}
	return nil
}